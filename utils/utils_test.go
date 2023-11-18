package utils

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/moby/buildkit/client/llb"
)

const (
	newDir       = "a/b/new_path"
	diffPermsDir = "a/diff_perms"
	existingDir  = "a/dir_exists"
	emptyFile    = "a/empty_file"
	nonemptyFile = "a/nonempty_file"

	// Note that we are using the /tmp folder, so use perms that
	// do not conflict with the sticky bit.
	testPerms = 0o711
)

var testRootDir string

func TestMain(m *testing.M) {
	var err error

	testRootDir, err = os.MkdirTemp("", "utils_test_*")
	if err != nil {
		log.Println("Failed to create test temp folder")
		return
	}

	defer func(p string) {
		_ = os.RemoveAll(p)
	}(testRootDir)

	testDir := path.Join(testRootDir, diffPermsDir)

	err = os.MkdirAll(testDir, 0o744)
	if err != nil {
		log.Printf("Failed to create test folder: %s\n", err)
		return
	}

	testDir = path.Join(testRootDir, existingDir)

	err = os.MkdirAll(testDir, testPerms)
	if err != nil {
		log.Printf("Failed to create test folder %s\n", testDir)
		return
	}

	testFile := path.Join(testRootDir, emptyFile)

	f, err := os.Create(testFile)
	if err != nil {
		log.Printf("Failed to create test file %s\n", testFile)
		return
	}

	_ = f.Close()

	testFile = path.Join(testRootDir, nonemptyFile)

	f, err = os.Create(testFile)
	if err != nil {
		log.Printf("Failed to create test file %s\n", testFile)
		return
	}

	_, err = f.WriteString("This is a non-empty test file")
	_ = f.Close()
	if err != nil {
		log.Printf("Failed to write to test file: %s\n", err)
		return
	}

	m.Run()
}

func TestEnsurePath(t *testing.T) {
	type args struct {
		path string
		perm os.FileMode
	}

	tests := []struct {
		name    string
		args    args
		created bool
		wantErr bool
	}{
		{"CreateNewPath", args{newDir, testPerms}, true, false},
		{"PathExists", args{existingDir, testPerms}, false, false},
		{"PathExistsWithDiffPerms", args{diffPermsDir, testPerms}, false, true},
		{"PathIsFile", args{emptyFile, testPerms}, false, true},
		{"EmptyPath", args{"", testPerms}, false, true},
		{"EmptyPerms", args{existingDir, 0o000}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := path.Join(testRootDir, tt.args.path)
			createdPath, err := EnsurePath(testPath, tt.args.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if createdPath != tt.created {
				t.Errorf("EnsurePath() created = %v, want %v", createdPath, tt.created)
			}
		})
	}

	_ = os.Remove(path.Join(testRootDir, newDir))
}

func TestIsNonEmptyFile(t *testing.T) {
	type args struct {
		dir  string
		file string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"NonEmptyFile", args{testRootDir, nonemptyFile}, true},
		{"EmptyFile", args{testRootDir, emptyFile}, false},
		{"MissingFile", args{testRootDir, "does_not_exist"}, false},
		{"UnspecifiedPath", args{"", existingDir}, false},
		{"UnspecifiedFile", args{testRootDir, ""}, false},
		{"PathIsDirectory", args{testRootDir, existingDir}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNonEmptyFile(tt.args.dir, tt.args.file); got != tt.want {
				t.Errorf("IsNonEmptyFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProxy(t *testing.T) {
	var got llb.ProxyEnv
	var want llb.ProxyEnv

	_ = os.Setenv("HTTP_PROXY", "httpproxy")
	_ = os.Setenv("HTTPS_PROXY", "httpsproxy")
	_ = os.Setenv("NO_PROXY", "noproxy")

	got = GetProxy()

	want = llb.ProxyEnv{
		HTTPProxy:  "httpproxy",
		HTTPSProxy: "httpsproxy",
		NoProxy:    "noproxy",
		AllProxy:   "httpproxy",
	}

	if got != want {
		t.Errorf("unexpected proxy config, got %#v want %#v", got, want)
	}

	_ = os.Unsetenv("HTTP_PROXY")
	_ = os.Unsetenv("HTTPS_PROXY")
	_ = os.Unsetenv("NO_PROXY")

	got = GetProxy()

	want = llb.ProxyEnv{
		HTTPProxy:  "",
		HTTPSProxy: "",
		NoProxy:    "",
		AllProxy:   "",
	}

	if got != want {
		t.Errorf("unexpected proxy config, got %#v want %#v", got, want)
	}
}
