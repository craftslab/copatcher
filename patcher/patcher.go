package patcher

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/craftslab/copatcher/config"
	"github.com/craftslab/copatcher/differ"
)

type Patcher interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, string) error
}

type Config struct {
	Config config.Config
	Differ *[]differ.Differ
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

func (p *patcher) Run(_ context.Context, name string) error {
	if _, err := os.Stat(name); err == nil {
		return errors.New("file exists already")
	}

	return nil
}
