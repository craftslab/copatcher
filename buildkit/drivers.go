package buildkit

import (
	"context"
	"fmt"
	"net/url"

	"github.com/moby/buildkit/client"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/apicaps"
	"github.com/pkg/errors"
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
