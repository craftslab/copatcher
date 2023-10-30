package buildkit

import (
	"context"

	"github.com/craftslab/copatcher/config"
)

const (
	DefaultAddr = "unix:///run/buildkit/buildkitd.sock"
)

type Buildkit interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
}

type buildkit struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Buildkit {
	return &buildkit{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (b *buildkit) Init(_ context.Context) error {
	return nil
}

func (b *buildkit) Deinit(_ context.Context) error {
	return nil
}

func (b *buildkit) Run(_ context.Context) error {
	return nil
}
