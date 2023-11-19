package utils

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

func EnsurePath(path string, perm fs.FileMode) (bool, error) {
	createdPath := false

	st, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, perm)
		createdPath = err == nil
	} else {
		if !st.IsDir() {
			return false, errors.New(fs.ErrExist.Error())
		}
		if st.Mode().Perm() != perm {
			return false, errors.New(fs.ErrPermission.Error())
		}
	}

	return createdPath, err
}

func IsNonEmptyFile(dir, file string) bool {
	p := filepath.Join(dir, file)

	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir() && info.Size() > 0
}

func GetProxy() llb.ProxyEnv {
	proxy := llb.ProxyEnv{
		HTTPProxy:  getEnvAny("HTTP_PROXY"),
		HTTPSProxy: getEnvAny("HTTPS_PROXY"),
		NoProxy:    getEnvAny("NO_PROXY"),
		AllProxy:   getEnvAny("HTTP_PROXY"),
	}

	return proxy
}

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}

	return ""
}
