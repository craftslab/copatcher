package report

import (
	"context"

	"github.com/craftslab/copatcher/config"
)

type Report interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
}

type report struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Report {
	return &report{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (r *report) Init(_ context.Context) error {
	return nil
}

func (r *report) Deinit(_ context.Context) error {
	return nil
}

func (r *report) Run(_ context.Context) error {
	return nil
}
