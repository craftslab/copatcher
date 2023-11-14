package cmd

import (
	"context"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/craftslab/copatcher/buildkit"
	"github.com/craftslab/copatcher/config"
	"github.com/craftslab/copatcher/patcher"
	"github.com/craftslab/copatcher/report"
)

var (
	app         = kingpin.New("copatcher", "Container patcher").Version(config.Version + "-build-" + config.Build)
	address     = app.Flag("address", "Address of buildkitd service").Default(buildkit.DefaultAddr).String()
	ignoreError = app.Flag("ignore-errors", "Ignore errors and continue patching").Bool()
	appImage    = app.Flag("image", "Application image name and tag to patch").Required().String()
	reportFile  = app.Flag("report", "Report file generated by container-diff").Required().String()
	tagName     = app.Flag("tag", "Tag for the patched image").Required().String()
	timeout     = app.Flag("timeout", "Timeout for the operation").Default(patcher.DefaultTimeout).String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cfg, err := initConfig(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	rp, err := initReport(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init report")
	}

	pt, err := initPatcher(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to init patcher")
	}

	if err := runPatcher(ctx, rp, pt); err != nil {
		return errors.Wrap(err, "failed to run patcher")
	}

	return nil
}

func initConfig(_ context.Context) (*config.Config, error) {
	c := config.New()
	return c, nil
}

func initReport(ctx context.Context, cfg *config.Config) (report.Report, error) {
	c := report.DefaultConfig()

	c.Config = *cfg

	return report.New(ctx, c), nil
}

func initPatcher(ctx context.Context, cfg *config.Config) (patcher.Patcher, error) {
	c := patcher.DefaultConfig()

	c.Config = *cfg

	return patcher.New(ctx, c), nil
}

func runPatcher(ctx context.Context, rp report.Report, pt patcher.Patcher) error {
	if err := pt.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	defer func(pt patcher.Patcher, ctx context.Context) {
		_ = pt.Deinit(ctx)
	}(pt, ctx)

	err := pt.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run")
	}

	return nil
}
