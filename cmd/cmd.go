package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/craftslab/dockerfiler/config"
	"github.com/craftslab/dockerfiler/differ"
	"github.com/craftslab/dockerfiler/filer"
)

var (
	app           = kingpin.New("dockerfiler", "Dockerfile generator").Version(config.Version + "-build-" + config.Build)
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

	f, err := initFiler(ctx, cfg, d)
	if err != nil {
		return errors.Wrap(err, "failed to init filer")
	}

	if err := runFiler(ctx, f); err != nil {
		return errors.Wrap(err, "failed to run filer")
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

func initFiler(ctx context.Context, cfg *config.Config, diff *[]differ.Differ) (filer.Filer, error) {
	c := filer.DefaultConfig()

	c.Config = *cfg
	c.Differ = diff

	return filer.New(ctx, c), nil
}

func runFiler(ctx context.Context, file filer.Filer) error {
	if err := file.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init")
	}

	defer func(file filer.Filer, ctx context.Context) {
		_ = file.Deinit(ctx)
	}(file, ctx)

	err := file.Run(ctx, *outputFile)
	if err != nil {
		return errors.Wrap(err, "failed to run")
	}

	return nil
}
