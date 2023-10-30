package patcher

import (
	"context"

	"github.com/craftslab/copatcher/config"
)

const (
	DefaultTimeout = "5m"
)

type Patcher interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
}

type patcher struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Patcher {
	return &patcher{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *patcher) Init(_ context.Context) error {
	return nil
}

func (p *patcher) Deinit(_ context.Context) error {
	return nil
}

func (p *patcher) Run(_ context.Context) error {
	return nil
}
