package cmd

import (
	"context"
	"os"

	"github.com/alecthomas/kingpin/v2"

	"github.com/craftslab/dockerfiler/config"
)

var (
	app        = kingpin.New("dockerfiler", "Dockerfile generator").Version(config.Version + "-build-" + config.Build)
	inputFile  = app.Flag("input-file", "Input file (.json)").Required().String()
	outputFile = app.Flag("output-file", "Output file (Dockerfile)").Required().String()
)

func Run(ctx context.Context) error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	return nil
}
