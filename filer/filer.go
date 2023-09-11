package filer

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/craftslab/dockerfiler/config"
	"github.com/craftslab/dockerfiler/differ"
)

type Filer interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, string) error
}

type Config struct {
	Config config.Config
	Differ *[]differ.Differ
}

type filer struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Filer {
	return &filer{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (f *filer) Init(_ context.Context) error {
	return nil
}

func (f *filer) Deinit(_ context.Context) error {
	return nil
}

func (f *filer) Run(_ context.Context, name string) error {
	if _, err := os.Stat(name); err == nil {
		return errors.New("file exists already")
	}

	return nil
}
