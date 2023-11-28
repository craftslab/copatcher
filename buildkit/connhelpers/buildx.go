package connhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cpuguy83/dockercfg"
	"github.com/cpuguy83/go-docker"
	"github.com/cpuguy83/go-docker/container"
	"github.com/cpuguy83/go-docker/errdefs"
	"github.com/moby/buildkit/client/connhelper"
	"github.com/pkg/errors"
)

// nolint: gochecknoinits
func init() {
	connhelper.Register("buildx", Buildx)
}

type buildxConfig struct {
	Driver string
	Nodes  []struct {
		Name     string
		Endpoint string
	}
}

// Buildx returns a buildkit connection helper for connecting to a buildx instance.
// Only "docker-container" buildkit instances are currently supported.
// If there are multiple nodes configured, one will be chosen at random.
func Buildx(u *url.URL) (*connhelper.ConnectionHelper, error) {
	if u.Path != "" {
		return nil, errors.Errorf("buildx driver does not support path elements: %s", u.Path)
	}

	return &connhelper.ConnectionHelper{
		ContextDialer: buildxContextDialer(u.Host),
	}, nil
}

func buildxContextDialer(builder string) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		configPath, err := dockercfg.ConfigPath()
		if err != nil {
			return nil, errors.Wrap(err, "failed to config path")
		}

		if builder == "" {
			// Standard env for setting a buildx builder name to use
			// This is used by buildx so we should use it too.
			builder = os.Getenv("BUILDX_BUILDER")
		}

		base := filepath.Join(filepath.Dir(configPath), "buildx")
		if builder == "" {
			dt, e := os.ReadFile(filepath.Join(base, "current"))
			if e != nil {
				return nil, errors.Wrap(e, "failed to read file")
			}
			type ref struct {
				Name string `json:"name"`
			}
			var r ref
			if e := json.Unmarshal(dt, &r); e != nil {
				return nil, errors.Wrap(e, "failed to unmarshal buildx config")
			}
			builder = r.Name
		}

		// Note: buildx inspect does not return json here, so we can't use the output directly
		cmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", "--bootstrap", builder)
		errBuf := bytes.NewBuffer(nil)
		cmd.Stderr = errBuf
		err = cmd.Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to inspect buildx instance")
		}

		// Read the config from the buildx instance
		dt, err := os.ReadFile(filepath.Join(base, "instances", builder))
		if err != nil {
			return nil, errors.Wrap(err, "failed to read file")
		}

		var cfg buildxConfig
		if err := json.Unmarshal(dt, &cfg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal buildx instance config")
		}

		if cfg.Driver != "docker-container" {
			return nil, errors.Errorf("unsupported buildx driver: %s", cfg.Driver)
		}

		if len(cfg.Nodes) == 0 {
			return nil, errors.New("no nodes configured for buildx instance")
		}

		nodes := cfg.Nodes
		if len(nodes) > 1 {
			rand.Shuffle(len(nodes), func(i, j int) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			})
		}

		return containerContextDialer(ctx, nodes[0].Endpoint, "buildx_buildkit_"+nodes[0].Name)
	}
}

func containerContextDialer(ctx context.Context, host, name string) (net.Conn, error) {
	tr, err := getDockerTransport(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get docker transport")
	}

	cli := docker.NewClient(docker.WithTransport(tr))
	c := cli.ContainerService().NewContainer(ctx, name)

	conn1, conn2 := net.Pipe()

	ep, err := c.Exec(ctx, container.WithExecCmd("buildctl", "dial-stdio"), func(cfg *container.ExecConfig) {
		cfg.Stdin = conn1
		cfg.Stdout = conn1
		cfg.Stderr = conn1
	})

	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, errors.Wrapf(err, "could not find container %s", name)
		}
		if err2 := c.Start(ctx); err2 != nil {
			return nil, errors.Wrap(err2, "failed to run start")
		}
		ep, err = c.Exec(ctx, container.WithExecCmd("buildctl", "dial-stdio"), func(cfg *container.ExecConfig) {
			cfg.Stdin = conn1
			cfg.Stdout = conn1
			cfg.Stderr = conn1
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to run exec")
		}
	}

	if err := ep.Start(ctx); err != nil {
		return nil, errors.Wrap(err, "could not start exec proxy")
	}

	return conn2, nil
}
