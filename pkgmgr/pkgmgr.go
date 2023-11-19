package pkgmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"

	"github.com/craftslab/copatcher/buildkit"
	"github.com/craftslab/copatcher/types"
)

const (
	copaPrefix     = "copa-"
	resultsPath    = "/" + copaPrefix + "out"
	downloadPath   = "/" + copaPrefix + "downloads"
	unpackPath     = "/" + copaPrefix + "unpacked"
	resultManifest = "results.manifest"
)

type PackageManager interface {
	InstallUpdates(context.Context, *types.UpdateManifest, bool) (*llb.State, []string, error)
	GetPackageType() string
}

func GetPackageManager(osType string, config *buildkit.Config, workingFolder string) (PackageManager, error) {
	switch osType {
	case "debian", "ubuntu":
		return &dpkgManager{config: config, workingFolder: workingFolder}, nil
	default:
		return nil, errors.New("unsupported OS type")
	}
}

// Utility functions for package manager implementations to share

type VersionComparer struct {
	IsValid  func(string) bool
	LessThan func(string, string) bool
}

// nolint: lll
func GetUniqueLatestUpdates(updates types.UpdatePackages, cmp VersionComparer, ignoreErrors bool) (types.UpdatePackages, error) {
	dict := make(map[string]string)
	var allErrors *multierror.Error

	for _, u := range updates {
		if cmp.IsValid(u.FixedVersion) {
			ver, ok := dict[u.Name]
			if !ok {
				dict[u.Name] = u.FixedVersion
			} else if cmp.LessThan(ver, u.FixedVersion) {
				dict[u.Name] = u.FixedVersion
			}
		} else {
			err := fmt.Errorf("invalid version %s found for package %s", u.FixedVersion, u.Name)
			allErrors = multierror.Append(allErrors, err)
			continue
		}
	}

	if allErrors != nil && !ignoreErrors {
		return types.UpdatePackages{}, allErrors.ErrorOrNil()
	}

	out := types.UpdatePackages{}
	for k, v := range dict {
		out = append(out, types.UpdatePackage{Name: k, FixedVersion: v})
	}

	return out, nil
}

type UpdatePackageInfo struct {
	Filename string
	Version  string
}

type PackageInfoReader interface {
	GetVersion(string) (string, error)
	GetName(string) (string, error)
}

type UpdateMap map[string]*UpdatePackageInfo

// nolint: lll
func GetValidatedUpdatesMap(updates types.UpdatePackages, cmp VersionComparer, reader PackageInfoReader, stagingPath string) (UpdateMap, error) {
	m := make(UpdateMap)

	for _, update := range updates {
		m[update.Name] = &UpdatePackageInfo{Version: update.FixedVersion}
	}

	files, err := os.ReadDir(stagingPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read directory")
	}

	if len(files) == 0 {
		return nil, nil
	}

	var allErrors *multierror.Error

	for _, file := range files {
		name, err := reader.GetName(file.Name())
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
			continue
		}
		version, err := reader.GetVersion(file.Name())
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
			continue
		}
		if !cmp.IsValid(version) {
			e := fmt.Errorf("invalid version %s found for package %s", version, name)
			allErrors = multierror.Append(allErrors, e)
			continue
		}
		p, ok := m[name]
		if !ok {
			_ = os.Remove(filepath.Join(stagingPath, file.Name()))
			continue
		}
		if cmp.LessThan(version, p.Version) {
			err = fmt.Errorf("downloaded package %s version %s lower than required %s for update", name, version, p.Version)
			allErrors = multierror.Append(allErrors, err)
			continue
		}
		p.Filename = file.Name()
	}

	if allErrors != nil {
		return nil, allErrors.ErrorOrNil()
	}

	return m, nil
}
