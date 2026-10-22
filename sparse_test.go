package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestSparseCheckoutCommand(t *testing.T) {
	cmd := sparseCheckout([]string{"src/backend", "dir with spaces", "-leading"})
	want := []string{"git", "sparse-checkout", "set", "--cone", "--", "src/backend", "dir with spaces", "-leading"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("args = %#v, want %#v", cmd.Args, want)
	}
}

func TestNormalizeSparseEntries(t *testing.T) {
	if got := normalizeSparseEntries([]string{"src", "", "src"}); !reflect.DeepEqual(got, []string{"src", "src"}) {
		t.Fatalf("normalization changed order or duplicates: %#v", got)
	}
}

func TestSparseFlags(t *testing.T) {
	t.Setenv("PLUGIN_SPARSE", "src/backend,shared")

	var paths []string
	command := &cli.Command{
		Flags: globalFlags,
		Action: func(_ context.Context, command *cli.Command) error {
			paths = command.StringSlice("sparse")
			return nil
		},
	}
	if err := command.Run(context.Background(), []string{"plugin-git"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{"src/backend", "shared"}) {
		t.Fatalf("sparse paths = %#v", paths)
	}
}

func TestSparseCheckoutLocalRepository(t *testing.T) {
	remote, commit := sparseFixture(t)
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "workspace")

	plugin := sparseFixturePlugin(remote, commit, home, workspace)
	plugin.Config.Sparse = []string{"src/backend", "shared"}
	if err := plugin.Exec(); err != nil {
		t.Fatal(err)
	}
	assertPathExists(t, workspace, "README.md", true)
	assertPathExists(t, workspace, "Directory.Build.props", true)
	assertPathExists(t, workspace, "src/backend/app.go", true)
	assertPathExists(t, workspace, "shared/common.go", true)
	assertPathExists(t, workspace, "src/frontend/app.ts", false)
	assertPathExists(t, workspace, "docs/guide.md", false)
	assertGitClean(t, workspace)

	// Reusing the same checkout with a different cone replaces the old paths.
	plugin.Config.Sparse = []string{"docs"}
	if err := plugin.Exec(); err != nil {
		t.Fatal(err)
	}
	assertPathExists(t, workspace, "docs/guide.md", true)
	assertPathExists(t, workspace, "src/backend/app.go", false)
	assertGitClean(t, workspace)

	// Omitting sparse checkout restores a complete cached workspace.
	plugin.Config.Sparse = nil
	if err := plugin.Exec(); err != nil {
		t.Fatal(err)
	}
	assertPathExists(t, workspace, "src/backend/app.go", true)
	assertPathExists(t, workspace, "src/frontend/app.ts", true)
	assertPathExists(t, workspace, "docs/guide.md", true)
	assertGitClean(t, workspace)
	if enabled, err := isSparseCheckoutEnabled(workspace, os.Environ()); err != nil || enabled {
		t.Fatalf("sparse checkout still enabled after full checkout: enabled=%v err=%v", enabled, err)
	}
}

func sparseFixturePlugin(remote, commit, home, workspace string) Plugin {
	return Plugin{
		Repo: Repo{Clone: remote},
		Pipeline: Pipeline{
			Path:   workspace,
			Commit: commit,
			Event:  "push",
		},
		Config: Config{
			Branch:    "main",
			Home:      home,
			Lfs:       false,
			Partial:   false,
			Recursive: false,
		},
	}
}

func sparseFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	remote := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "init", "-b", "main")
	runGit(t, source, "config", "user.name", "Sparse Test")
	runGit(t, source, "config", "user.email", "sparse@example.invalid")
	files := map[string]string{
		"README.md":                      "root\n",
		"Directory.Build.props":          "props\n",
		"src/backend/app.go":             "package backend\n",
		"src/backend/nested/config.json": "{}\n",
		"src/frontend/app.ts":            "export {}\n",
		"shared/common.go":               "package shared\n",
		"docs/guide.md":                  "guide\n",
		"dir with spaces/file.txt":       "spaces\n",
	}
	for name, contents := range files {
		path := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "fixture")
	commit := gitOutput(t, source, "rev-parse", "HEAD")
	runGit(t, root, "clone", "--bare", source, remote)
	return filepath.ToSlash(remote), commit
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func assertPathExists(t *testing.T, root, path string, want bool) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	got := err == nil
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("path %q exists = %v, want %v", path, got, want)
	}
}

func assertGitClean(t *testing.T, workspace string) {
	t.Helper()
	if got := gitOutput(t, workspace, "status", "--porcelain"); got != "" {
		t.Fatalf("working tree is not clean:\n%s", got)
	}
}
