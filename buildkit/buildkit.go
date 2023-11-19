package buildkit

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/containerd/console"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/buildx/build"
	"github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/contentutil"
	"github.com/moby/buildkit/util/imageutil"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/version"
	"github.com/opencontainers/go-digest"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/craftslab/copatcher/types"
)

type Config struct {
	ImageName  string
	Client     *client.Client
	ConfigData []byte
	Platform   ispec.Platform
	ImageState llb.State
}

type Opts struct {
	Addr       string
	CACertPath string
	CertPath   string
	KeyPath    string
}

// nolint: lll
func InitializeBuildkitConfig(ctx context.Context, clt *client.Client, image string, manifest *types.UpdateManifest) (*Config, error) {
	// Initialize buildkit config for the target image
	cfg := Config{
		ImageName: image,
		Platform: ispec.Platform{
			OS:           "linux",
			Architecture: manifest.Metadata.Config.Arch,
		},
	}

	// Resolve and pull the config for the target image
	_, configData, err := resolveImageConfig(ctx, image, &cfg.Platform)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve image config")
	}

	cfg.ConfigData = configData

	// Load the target image state with the resolved image config in case environment variable settings
	// are necessary for running apps in the target image for updates
	cfg.ImageState, err = llb.Image(image,
		llb.Platform(cfg.Platform),
		llb.ResolveModeDefault,
	).WithImageConfig(cfg.ConfigData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load image")
	}

	cfg.Client = clt

	return &cfg, nil
}

func SolveToLocal(ctx context.Context, c *client.Client, st *llb.State, outPath string) error {
	def, err := st.Marshal(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run marshal")
	}

	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(dockerConfig)}
	solveOpt := client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type:      client.ExporterLocal,
				OutputDir: outPath,
			},
		},
		Frontend: "",         // i.e. we are passing in the llb.Definition directly
		Session:  attachable, // used for authprovider, sshagentprovider and secretprovider
	}

	solveOpt.SourcePolicy, err = build.ReadSourcePolicy()
	if err != nil {
		return errors.Wrap(err, "failed to read source policy")
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, e := c.Solve(ctx, def, solveOpt, ch)
		return e
	})

	eg.Go(func() error {
		var c console.Console
		cn, e := console.ConsoleFromFile(os.Stderr)
		if e == nil {
			c = cn
		}
		// not using shared context to not disrupt display but let us finish reporting errors
		_, e = progressui.DisplaySolveStatus(context.TODO(), c, os.Stdout, ch)
		return e
	})

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to run wait")
	}

	return nil
}

func SolveToDocker(ctx context.Context, c *client.Client, st *llb.State, configData []byte, tag string) error {
	def, err := st.Marshal(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run marshal")
	}

	pipeR, pipeW := io.Pipe()
	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(dockerConfig)}

	solveOpt := client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterDocker,
				Attrs: map[string]string{
					"name": tag,
					// Pass through resolved configData from original image
					exptypes.ExporterImageConfigKey: string(configData),
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return pipeW, nil
				},
			},
		},
		Frontend: "",         // i.e. we are passing in the llb.Definition directly
		Session:  attachable, // used for authprovider, sshagentprovider and secretprovider
	}

	solveOpt.SourcePolicy, err = build.ReadSourcePolicy()
	if err != nil {
		return errors.Wrap(err, "failed to read source policy")
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, e := c.Solve(ctx, def, solveOpt, ch)
		return errors.Wrap(e, "failed to run solve")
	})

	eg.Go(func() error {
		var c console.Console
		cn, e := console.ConsoleFromFile(os.Stderr)
		if e == nil {
			c = cn
		}
		// not using shared context to not disrupt display but let us finish reporting errors
		_, e = progressui.DisplaySolveStatus(context.TODO(), c, os.Stdout, ch)
		return errors.Wrap(e, "failed to display solve status")
	})

	eg.Go(func() error {
		if err := dockerLoad(ctx, pipeR); err != nil {
			return errors.Wrap(err, "failed to load docker")
		}
		return pipeR.Close()
	})

	return eg.Wait()
}

func dockerLoad(ctx context.Context, pipeR io.Reader) error {
	cmd := exec.CommandContext(ctx, "docker", "load")
	cmd.Stdin = pipeR

	_, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to run stdout pipe")
	}

	_, err = cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "failed to run stderr pipe")
	}

	return cmd.Run()
}

// Custom ResolveImageConfig implementation for using Docker default config.json credentials
// to pull image config.
//
// While it would be ideal to be able to use imagemetaresolver.Default().ResolveImageConfig(),
// there doesn't seem to be a way to configure the necessary DockerAuthorizer or RegistryHosts
// against an ImageMetaResolver, which causes the resolve to only use anonymous tokens and fail.
func resolveImageConfig(ctx context.Context, ref string, platform *ispec.Platform) (digest.Digest, []byte, error) {
	auth := docker.NewDockerAuthorizer(
		docker.WithAuthCreds(func(ref string) (string, string, error) {
			defaultConfig := config.LoadDefaultConfigFile(os.Stderr)
			ac, err := defaultConfig.GetAuthConfig(ref)
			if err != nil {
				return "", "", errors.Wrap(err, "failed to get auth config")
			}
			if ac.IdentityToken != "" {
				return "", ac.IdentityToken, nil
			}
			return ac.Username, ac.Password, nil
		}))

	hosts := docker.ConfigureDefaultRegistries(
		docker.WithClient(http.DefaultClient),
		docker.WithPlainHTTP(docker.MatchLocalhost),
		docker.WithAuthorizer(auth),
	)

	headers := http.Header{}
	headers.Set("User-Agent", version.UserAgent())

	resolver := docker.NewResolver(docker.ResolverOptions{
		Client:  http.DefaultClient,
		Headers: headers,
		Hosts:   hosts,
	})

	_, dgst, cfg, err := imageutil.Config(ctx, ref, resolver, contentutil.NewBuffer(), nil, platform, nil)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to run config")
	}

	return dgst, cfg, nil
}
