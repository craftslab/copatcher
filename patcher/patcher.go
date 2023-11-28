package patcher

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/buildkit/client"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"github.com/craftslab/copatcher/buildkit"
	"github.com/craftslab/copatcher/config"
	"github.com/craftslab/copatcher/pkgmgr"
	"github.com/craftslab/copatcher/report"
	"github.com/craftslab/copatcher/utils"
)

const (
	DefaultFolder  = "/tmp/copatcher"
	DefaultPerm    = 0o744
	DefaultTag     = "patched"
	DefaultTimeout = "5m"
)

type Patcher interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, string) error
}

type Config struct {
	Config       config.Config
	IgnoreErrors bool
	Image        string
	Report       report.Report
	Tag          string
	Timeout      time.Duration
}

type patcher struct {
	cfg  *Config
	opts buildkit.Opts
}

func New(_ context.Context, cfg *Config) Patcher {
	return &patcher{
		cfg: cfg,
	}
}

func DefaultConfig() *Config {
	return &Config{}
}

func (p *patcher) Init(ctx context.Context) error {
	if err := p.cfg.Report.Init(ctx); err != nil {
		return errors.Wrap(err, "failed to init report")
	}

	return nil
}

func (p *patcher) Deinit(ctx context.Context) error {
	_ = p.cfg.Report.Deinit(ctx)

	return nil
}

func (p *patcher) Run(ctx context.Context, name string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	ch := make(chan error)
	go func() {
		ch <- p.patch(timeoutCtx, name)
	}()

	select {
	case err := <-ch:
		return errors.Wrap(err, "failed to patch")
	case <-timeoutCtx.Done():
		<-time.After(1 * time.Second)
		return errors.New("patch exceeded timeout")
	}
}

// nolint: funlen,gocyclo
func (p *patcher) patch(ctx context.Context, name string) error {
	imageName, err := reference.ParseNamed(p.cfg.Image)
	if err != nil {
		return errors.Wrap(err, "failed to parse named")
	}

	if reference.IsNameOnly(imageName) {
		imageName = reference.TagNameOnly(imageName)
	}

	taggedName, ok := imageName.(reference.Tagged)
	if !ok {
		return errors.New("invalid tagged name")
	}

	tag := taggedName.Tag()
	if p.cfg.Tag == "" {
		if tag == "" {
			p.cfg.Tag = DefaultTag
		} else {
			p.cfg.Tag = fmt.Sprintf("%s-%s", tag, DefaultTag)
		}
	}

	patchedImageName := fmt.Sprintf("%s:%s", imageName.Name(), p.cfg.Tag)

	if isNew, e := utils.EnsurePath(DefaultFolder, DefaultPerm); e != nil {
		return errors.Wrap(e, "failed to create working folder")
	} else if isNew {
		defer func(p string) {
			_ = os.RemoveAll(p)
		}(DefaultFolder)
	}

	manifest, err := p.cfg.Report.Run(ctx, name)
	if err != nil {
		return errors.Wrap(err, "failed to parse report")
	}

	_client, err := buildkit.NewClient(ctx, p.opts)
	if err != nil {
		return errors.Wrap(err, "failed to create new client")
	}

	defer func(c *client.Client) {
		_ = c.Close()
	}(_client)

	_config, err := buildkit.InitializeBuildkitConfig(ctx, _client, p.cfg.Image, &manifest)
	if err != nil {
		return errors.Wrap(err, "failed to init buildkit config")
	}

	_pkgmgr, err := pkgmgr.GetPackageManager(manifest.Metadata.OS.Type, _config, DefaultFolder)
	if err != nil {
		return errors.Wrap(err, "failed to get package manager")
	}

	patchedImageState, errPkgs, err := _pkgmgr.InstallUpdates(ctx, &manifest, p.cfg.IgnoreErrors)
	if err != nil {
		return errors.Wrap(err, "failed to install updates")
	}

	if err := buildkit.SolveToDocker(ctx, _config.Client, patchedImageState, _config.ConfigData, patchedImageName); err != nil {
		return errors.Wrap(err, "failed to solve to docker")
	}

	for _, update := range manifest.Updates {
		if !slices.Contains(errPkgs, update.Name) {
			// TODO: FIXME
		}
	}

	return nil
}
