package patcher

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/buildkit/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/craftslab/copatcher/buildkit"
	"github.com/craftslab/copatcher/config"
	"github.com/craftslab/copatcher/pkgmgr"
	"github.com/craftslab/copatcher/types"
	"github.com/craftslab/copatcher/utils"
)

const (
	DefaultTimeout    = "5m"
	DefaultPatchedTag = "patched"
)

const (
	FolderPerm = 0o744
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

func Patch(ctx context.Context, timeout time.Duration, image, patchedTag, workingFolder string,
	manifest *types.UpdateManifest, ignoreError bool, bkOpts buildkit.Opts) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ch := make(chan error)
	go func() {
		ch <- patchWithContext(timeoutCtx, image, patchedTag, workingFolder, manifest, ignoreError, bkOpts)
	}()

	select {
	case err := <-ch:
		return err
	case <-timeoutCtx.Done():
		<-time.After(1 * time.Second)
		err := fmt.Errorf("patch exceeded timeout %v", timeout)
		log.Error(err)
		return err
	}
}

// nolint: funlen,gocyclo
func patchWithContext(ctx context.Context, image, patchedTag, workingFolder string, manifest *types.UpdateManifest,
	ignoreError bool, bkOpts buildkit.Opts) error {
	imageName, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}

	if reference.IsNameOnly(imageName) {
		log.Warnf("Image name has no tag or digest, using latest as tag")
		imageName = reference.TagNameOnly(imageName)
	}

	taggedName, ok := imageName.(reference.Tagged)
	if !ok {
		e := errors.New("unexpected: TagNameOnly did create Tagged ref")
		log.Error(e)
		return e
	}

	tag := taggedName.Tag()
	if patchedTag == "" {
		if tag == "" {
			log.Warnf("No output tag specified for digest-referenced image, defaulting to `%s`", DefaultPatchedTag)
			patchedTag = DefaultPatchedTag
		} else {
			patchedTag = fmt.Sprintf("%s-%s", tag, DefaultPatchedTag)
		}
	}

	patchedImageName := fmt.Sprintf("%s:%s", imageName.Name(), patchedTag)

	// Ensure working folder exists for call to InstallUpdates
	if workingFolder == "" {
		workingFolder, err = os.MkdirTemp("", "copa-*")
		if err != nil {
			return err
		}
		defer func(p string) {
			_ = os.RemoveAll(p)
		}(workingFolder)
		if e := os.Chmod(workingFolder, FolderPerm); e != nil {
			return e
		}
	} else {
		if isNew, e := utils.EnsurePath(workingFolder, FolderPerm); e != nil {
			log.Errorf("failed to create workingFolder %s", workingFolder)
			return e
		} else if isNew {
			defer func(p string) {
				_ = os.RemoveAll(p)
			}(workingFolder)
		}
	}

	_client, err := buildkit.NewClient(ctx, bkOpts)
	if err != nil {
		return err
	}

	defer func(c *client.Client) {
		_ = c.Close()
	}(_client)

	// Configure buildctl/client for use by package manager
	_config, err := buildkit.InitializeBuildkitConfig(ctx, _client, image, manifest)
	if err != nil {
		return err
	}

	// Create package manager helper
	_pkgmgr, err := pkgmgr.GetPackageManager(manifest.Metadata.OS.Type, _config, workingFolder)
	if err != nil {
		return err
	}

	// Export the patched image state to Docker
	// TODO: Add support for other output modes as buildctl does.
	patchedImageState, errPkgs, err := _pkgmgr.InstallUpdates(ctx, manifest, ignoreError)
	if err != nil {
		return err
	}

	if err := buildkit.SolveToDocker(ctx, _config.Client, patchedImageState, _config.ConfigData, patchedImageName); err != nil {
		return err
	}

	// create a new manifest with the successfully patched packages
	validatedManifest := &types.UpdateManifest{
		Metadata: types.Metadata{
			OS: types.OS{
				Type:    manifest.Metadata.OS.Type,
				Version: manifest.Metadata.OS.Version,
			},
			Config: types.Config{
				Arch: manifest.Metadata.Config.Arch,
			},
		},
		Updates: []types.UpdatePackage{},
	}

	for _, update := range manifest.Updates {
		if !slices.Contains(errPkgs, update.Name) {
			validatedManifest.Updates = append(validatedManifest.Updates, update)
		}
	}

	return nil
}
