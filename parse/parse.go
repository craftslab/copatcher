package parse

import (
	"context"

	"github.com/craftslab/copatcher/config"
)

type Parse interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
}

type parse struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Parse {
	return &parse{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *parse) Init(_ context.Context) error {
	return nil
}

func (p *parse) Deinit(_ context.Context) error {
	return nil
}

func (p *parse) Run(_ context.Context) error {
	return nil
}
