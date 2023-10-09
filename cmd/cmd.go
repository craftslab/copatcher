package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/craftslab/copatcher/config"
	"github.com/craftslab/copatcher/differ"
	"github.com/craftslab/copatcher/patcher"
)

var (
	app           = kingpin.New("copatcher", "Container patcher").Version(config.Version + "-build-" + config.Build)
	containerDiff = app.Flag("container-diff", "Container difference (.json)").Required().String()
	outputFile    = app.Flag("output-file", "Output file (Dockerfile)").Required().String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	cfg, err := initConfig(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	d, err := initDiffer(ctx, cfg, *containerDiff)
	if err != nil {
		return errors.Wrap(err, "failed to init differ")
	}

	p, err := initPatcher(ctx, cfg, d)
	if err != nil {
		return errors.Wrap(err, "failed to init patcher")
	}

	if err := runPatcher(ctx, p); err != nil {
		return errors.Wrap(err, "failed to run patcher")
	}

	return nil
}

func initConfig(_ context.Context) (*config.Config, error) {
	c := config.New()
	return c, nil
}

func initDiffer(_ context.Context, _ *config.Config, name string) (*[]differ.Differ, error) {
	d := differ.New()

	fi, err := os.Open(name)
	if err != nil {
		return d, errors.Wrap(err, "failed to open")
	}

	defer func() {
		_ = fi.Close()
	}()

	buf, _ := io.ReadAll(fi)

	if err := json.Unmarshal(buf, d); err != nil {
		return d, errors.Wrap(err, "failed to unmarshal")
	}

	return d, nil
}

func initPatcher(ctx context.Context, cfg *config.Config, diff *[]differ.Differ) (patcher.Patcher, error) {
	c := patcher.DefaultConfig()

	c.Config = *cfg
	c.Differ = diff

	return patcher.New(ctx, c), nil
}

func runPatcher(ctx context.Context, pat patcher.Patcher) error {
	if err := pat.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	defer func(pat patcher.Patcher, ctx context.Context) {
		_ = pat.Deinit(ctx)
	}(pat, ctx)

	err := pat.Run(ctx, *outputFile)
	if err != nil {
		return errors.Wrap(err, "failed to run")
	}

	return nil
}
