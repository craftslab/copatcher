package buildkit

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/moby/buildkit/client"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/apicaps"
	"github.com/pkg/errors"

	"github.com/craftslab/copatcher/buildkit/connhelpers"
)

const (
	DefaultAddr = "unix:///run/buildkit/buildkitd.sock"
)

var (
	errMissingCap = fmt.Errorf("missing required buildkit functionality")
	// requiredCaps are buildkit llb ops required to function.
	requiredCaps = []apicaps.CapID{pb.CapMergeOp, pb.CapDiffOp}
)

// NewClient returns a new buildkit client with the given addr.
// If addr is empty it will first try to connect to docker's buildkit instance and then fallback to DefaultAddr.
func NewClient(ctx context.Context, bkOpts Opts) (*client.Client, error) {
	if bkOpts.Addr == "" {
		return autoClient(ctx)
	}

	opts := getCredentialOptions(bkOpts)

	clt, err := client.New(ctx, bkOpts.Addr, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run new")
	}

	return clt, nil
}

func getCredentialOptions(bkOpts Opts) []client.ClientOpt {
	opts := []client.ClientOpt{}
	if bkOpts.CACertPath != "" {
		opts = append(opts, client.WithServerConfig(getServerNameFromAddr(bkOpts.Addr), bkOpts.CACertPath))
	}

	if bkOpts.CertPath != "" || bkOpts.KeyPath != "" {
		opts = append(opts, client.WithCredentials(bkOpts.CertPath, bkOpts.KeyPath))
	}

	return opts
}

func getServerNameFromAddr(addr string) string {
	u, err := url.Parse(addr)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// ValidateClient checks to ensure the connected buildkit instance supports the features required by copa.
func ValidateClient(ctx context.Context, c *client.Client) error {
	_, err := c.Build(ctx, client.SolveOpt{}, "", func(ctx context.Context, client gateway.Client) (*gateway.Result, error) {
		capset := client.BuildOpts().LLBCaps
		var err error
		for _, _cap := range requiredCaps {
			err = errors.Wrap(err, capset.Supports(_cap).Error())
		}
		if err != nil {
			return nil, errors.Wrap(err, errMissingCap.Error())
		}
		return &gateway.Result{}, nil
	}, nil)

	return err
}

func autoClient(ctx context.Context, opts ...client.ClientOpt) (*client.Client, error) {
	var retErr error

	newClient := func(ctx context.Context, dialer func(context.Context, string) (net.Conn, error)) (*client.Client, error) {
		_client, err := client.New(ctx, "", append(opts, client.WithContextDialer(dialer))...)
		if err == nil {
			err = ValidateClient(ctx, _client)
			if err == nil {
				return _client, nil
			}
			_ = _client.Close()
		}
		return nil, err
	}

	h, err := connhelpers.Docker(&url.URL{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to run docker")
	}

	c, err := newClient(ctx, h.ContextDialer)
	if err == nil {
		return c, nil
	}

	retErr = errors.Wrap(retErr, "could not use docker driver")

	h, err = connhelpers.Buildx(&url.URL{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to run buildx")
	}

	c, err = newClient(ctx, h.ContextDialer)
	if err == nil {
		return c, nil
	}
	retErr = errors.Wrap(retErr, "could not use buildx driver")

	c, err = client.New(ctx, DefaultAddr, opts...)
	if err == nil {
		err = ValidateClient(ctx, c)
		if err == nil {
			return c, nil
		}
		_ = c.Close()
	}

	return nil, errors.Wrap(retErr, "could not use buildkitd driver")
}
