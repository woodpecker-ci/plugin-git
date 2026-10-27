package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lfsMergeFixture creates a remote where "main" gains an LFS file under
// src/backend after the PR branch "feature" diverged from it.
func lfsMergeFixture(t *testing.T) (remote, prCommit string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	remote = filepath.Join(root, "remote.git")
	write := func(name, contents string) {
		path := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "init", "--bare", "-b", "main", remote)
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "init", "-b", "main")
	runGit(t, source, "config", "user.name", "Sparse Test")
	runGit(t, source, "config", "user.email", "sparse@example.invalid")
	runGit(t, source, "lfs", "install", "--local")
	runGit(t, source, "remote", "add", "origin", remote)
	write(".gitattributes", "*.bin filter=lfs diff=lfs merge=lfs -text\n")
	write("src/backend/app.go", "package backend\n")
	write("src/backend/base.bin", "base lfs content\n")
	write("src/frontend/app.ts", "export {}\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "base")

	runGit(t, source, "switch", "-c", "feature")
	write("src/backend/feature.go", "package backend\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "feature")
	prCommit = gitOutput(t, source, "rev-parse", "HEAD")

	runGit(t, source, "switch", "main")
	write("src/backend/merged.bin", "lfs content from target branch\n")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "target adds lfs file")

	runGit(t, source, "push", "origin", "main", "feature")
	return filepath.ToSlash(remote), prCommit
}

func runMergePR(t *testing.T, sparse []string) string {
	remote, commit := lfsMergeFixture(t)
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "workspace")
	plugin := sparseFixturePlugin(remote, commit, home, workspace)
	plugin.Pipeline.Event = "pull_request"
	plugin.Config.Event = "pull_request"
	plugin.Config.Branch = "feature"
	plugin.Config.Lfs = true
	plugin.Config.MergePullRequest = true
	plugin.Config.TargetBranch = "main"
	plugin.Config.GitUserName = "CI"
	plugin.Config.GitUserEmail = "ci@example.invalid"
	plugin.Config.Sparse = sparse
	if err := plugin.Exec(); err != nil {
		t.Fatal(err)
	}
	return workspace
}

func assertLFSMaterialized(t *testing.T, workspace, path, want string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(got), "version https://git-lfs") {
		t.Errorf("%s is still an LFS pointer:\n%s", path, got)
	} else if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func TestReproLFSMergeFull(t *testing.T) {
	ws := runMergePR(t, nil)
	assertLFSMaterialized(t, ws, "src/backend/base.bin", "base lfs content\n")
	assertLFSMaterialized(t, ws, "src/backend/merged.bin", "lfs content from target branch\n")
}

func TestReproLFSMergeSparse(t *testing.T) {
	ws := runMergePR(t, []string{"src/backend"})
	assertPathExists(t, ws, "src/frontend/app.ts", false)
	assertLFSMaterialized(t, ws, "src/backend/base.bin", "base lfs content\n")
	assertLFSMaterialized(t, ws, "src/backend/merged.bin", "lfs content from target branch\n")
}
