package pkgmgr

import (
	"context"

	"github.com/craftslab/copatcher/config"
)

type Pkgmgr interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context) error
}

type Config struct {
	Config config.Config
}

type pkgmgr struct {
	cfg *Config
}

func New(_ context.Context, cfg *Config) Pkgmgr {
	return &pkgmgr{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *pkgmgr) Init(_ context.Context) error {
	return nil
}

func (p *pkgmgr) Deinit(_ context.Context) error {
	return nil
}

func (p *pkgmgr) Run(_ context.Context) error {
	return nil
}
