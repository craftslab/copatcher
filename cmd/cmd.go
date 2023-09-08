package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/craftslab/dockerfiler/diff"
)

var (
	app           = kingpin.New("dockerfiler", "Dockerfile generator").Version(diff.Version + "-build-" + diff.Build)
	containerDiff = app.Flag("container-diff", "Container difference (.json)").Required().String()
	outputFile    = app.Flag("output-file", "Output file (Dockerfile)").Required().String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	_, err := initDiff(ctx, *containerDiff)
	if err != nil {
		return errors.Wrap(err, "failed to init diff")
	}

	if _, err := os.Stat(*outputFile); err == nil {
		return errors.Wrap(err, "file exists already")
	}

	return nil
}

func initDiff(_ context.Context, name string) (*diff.Diff, error) {
	d := diff.New()

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
