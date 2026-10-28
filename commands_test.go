package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
)

const testCommit = "0123456789012345678901234567890123456789"

// TestCommands checks the git commands built for a new repository, without
// executing them. Commands that let git-lfs materialize LFS files are prefixed
// with "[smudge]".
func TestCommands(t *testing.T) {
	initCmds := []string{
		"git init --object-format sha1 -b main",
		"git config --global --replace-all safe.directory /workspace",
		"git remote add origin https://example.com/repo.git",
	}
	fetchCommit := "git fetch --no-tags origin +" + testCommit + ":"
	checkoutCommit := "git reset --hard -q " + testCommit

	testdata := []struct {
		name   string
		config func(p *Plugin)
		want   []string
		err    string
	}{
		{
			name:   "commit",
			config: func(p *Plugin) {},
			want:   slices.Concat(initCmds, []string{fetchCommit, checkoutCommit}),
		},
		{
			name:   "lfs",
			config: func(p *Plugin) { p.Config.Lfs = true },
			want: slices.Concat(initCmds, []string{
				fetchCommit,
				checkoutCommit,
				"git lfs fetch",
				"git lfs checkout",
			}),
		},
		{
			name:   "ref",
			config: func(p *Plugin) { p.Pipeline.Ref = "refs/pull/1/head" },
			want: slices.Concat(initCmds, []string{
				"git fetch --no-tags origin +refs/pull/1/head:",
				"git checkout -qf FETCH_HEAD",
			}),
		},
		{
			name:   "short sha",
			config: func(p *Plugin) { p.Pipeline.Commit = "abc123" },
			want: slices.Concat(initCmds, []string{
				"git fetch --no-tags origin +main:",
				"git switch -q main",
				"git reset --hard -q abc123",
			}),
		},
		{
			name: "short sha without branch",
			config: func(p *Plugin) {
				p.Pipeline.Commit = "abc123"
				p.Config.Branch = ""
			},
			err: "short commit SHA1 checkout requires a branch",
		},
		{
			name:   "sha256",
			config: func(p *Plugin) { p.Pipeline.Commit = testCommit + testCommit[:24] },
			want: []string{
				"git init --object-format sha256 -b main",
				initCmds[1],
				initCmds[2],
				"git fetch --no-tags origin +" + testCommit + testCommit[:24] + ":",
				"git reset --hard -q " + testCommit + testCommit[:24],
			},
		},
		{
			name: "skip verify takes precedence over custom cert",
			config: func(p *Plugin) {
				p.Config.SkipVerify = true
				p.Config.CustomCert = "/cert.pem"
			},
			want: slices.Concat([]string{"git config --global http.sslVerify false"}, initCmds, []string{fetchCommit, checkoutCommit}),
		},
		{
			name:   "custom cert",
			config: func(p *Plugin) { p.Config.CustomCert = "/cert.pem" },
			want:   slices.Concat([]string{"git config --global http.sslCAInfo /cert.pem"}, initCmds, []string{fetchCommit, checkoutCommit}),
		},
		{
			name: "ssh",
			config: func(p *Plugin) {
				p.Config.UseSSH = true
				p.Config.SSHKey = "/sshkey"
			},
			want: slices.Concat(initCmds[:2], []string{
				"git remote add origin git@example.com:repo.git",
				`git config --global core.sshCommand ssh -i /sshkey -o "StrictHostKeyChecking no"`,
				fetchCommit,
				checkoutCommit,
			}),
		},
		{
			name: "ssh with private key and host key",
			config: func(p *Plugin) {
				p.Config.UseSSH = true
				p.Config.SSHKeyPrivate = "private key"
				p.Config.SSHHostKey = "host key"
			},
			want: slices.Concat(initCmds[:2], []string{
				"git remote add origin git@example.com:repo.git",
				"git config --global core.sshCommand ssh -i " + filepath.Join("$HOME", "sshkey") +
					" -o UserKnownHostsFile=" + filepath.Join("$HOME", ".ssh", "known_hosts"),
				fetchCommit,
				checkoutCommit,
			}),
		},
		{
			name: "submodules",
			config: func(p *Plugin) {
				p.Config.Submodules = `{"lib":"https://example.com/lib.git"}`
				p.Config.Recursive = true
				p.Config.SubmodulePartial = true
			},
			want: slices.Concat(initCmds, []string{
				fetchCommit,
				checkoutCommit,
				"git config --global submodule.lib.url https://example.com/lib.git",
				"git submodule update --init --recursive --depth=1 --recommend-shallow",
			}),
		},
		{
			name:   "invalid submodule overrides",
			config: func(p *Plugin) { p.Config.Submodules = `{` },
			err:    "could not parse submodule_override map: unexpected end of JSON input",
		},
		{
			name: "pull request merge",
			config: func(p *Plugin) {
				p.Config.Event = "pull_request"
				p.Config.MergePullRequest = true
				p.Config.TargetBranch = "main"
				p.Config.GitUserName = "ci"
				p.Config.GitUserEmail = "ci@example.com"
				p.Config.Depth = 50
			},
			want: slices.Concat(initCmds, []string{
				fetchCommit, // depth is ignored to be able to merge
				checkoutCommit,
				"git config --global user.name ci",
				"git config --global user.email ci@example.com",
				"git fetch origin -- main",
				"git merge origin/main",
			}),
		},
		{
			name: "pull request merge with sparse lfs",
			config: func(p *Plugin) {
				p.Config.Event = "pull_request"
				p.Config.MergePullRequest = true
				p.Config.TargetBranch = "main"
				p.Config.GitUserName = "ci"
				p.Config.GitUserEmail = "ci@example.com"
				p.Config.Lfs = true
				p.Config.Sparse = []string{"src", ""}
			},
			want: slices.Concat(initCmds, []string{
				fetchCommit,
				"git lfs install --local",
				"[smudge] git sparse-checkout set --cone -- src",
				"[smudge] " + checkoutCommit,
				"git config --global user.name ci",
				"git config --global user.email ci@example.com",
				"git fetch origin -- main",
				"[smudge] git merge origin/main",
			}),
		},
		{
			name: "pull request fetch target branch",
			config: func(p *Plugin) {
				p.Config.Event = "pull_request"
				p.Config.FetchTargetBranch = true
				p.Config.TargetBranch = "main"
			},
			want: slices.Concat(initCmds, []string{fetchCommit, checkoutCommit, "git fetch origin -- main"}),
		},
		{
			name: "pull request fetch without target branch",
			config: func(p *Plugin) {
				p.Config.Event = "pull_request"
				p.Config.FetchTargetBranch = true
			},
			err: "fetch_target_branch is enabled but target_branch is not set",
		},
		{
			name: "pull request settings ignored on push",
			config: func(p *Plugin) {
				p.Config.MergePullRequest = true
				p.Config.FetchTargetBranch = true
				p.Config.TargetBranch = "main"
			},
			want: slices.Concat(initCmds, []string{fetchCommit, checkoutCommit}),
		},
		{
			name: "sparse",
			config: func(p *Plugin) {
				p.Config.Sparse = []string{"src", "docs"}
				p.Config.Depth = 1
				p.Config.filter = "blob:none"
			},
			want: slices.Concat(initCmds, []string{
				"git fetch --no-tags --depth=1 --filter=blob:none origin +" + testCommit + ":",
				"git sparse-checkout set --cone -- src docs",
				checkoutCommit,
			}),
		},
	}

	for _, td := range testdata {
		t.Run(td.name, func(t *testing.T) {
			home := t.TempDir()
			p := Plugin{
				Repo: Repo{
					Clone:    "https://example.com/repo.git",
					CloneSSH: "git@example.com:repo.git",
				},
				Pipeline: Pipeline{Path: t.TempDir(), Commit: testCommit, Event: "push"},
				Config: Config{
					Branch:        "main",
					Home:          home,
					Event:         "push",
					SafeDirectory: "/workspace",
				},
			}
			td.config(&p)
			p.normalizeConfig()

			cmds, err := p.commands(nil)
			if td.err != "" {
				if err == nil || err.Error() != td.err {
					t.Fatalf("error = %v, want %q", err, td.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			got := make([]string, 0, len(cmds))
			for _, cmd := range cmds {
				line := strings.ReplaceAll(strings.Join(cmd.Args, " "), home, "$HOME")
				if !slices.Contains(cmd.Env, "GIT_LFS_SKIP_SMUDGE=1") {
					line = "[smudge] " + line
				}
				got = append(got, line)
			}
			if !reflect.DeepEqual(got, td.want) {
				t.Errorf("commands:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(td.want, "\n"))
			}
		})
	}
}

func TestSSHSetupWritesKeys(t *testing.T) {
	home := t.TempDir()
	p := Plugin{Config: Config{
		Home:          home,
		SSHKeyPrivate: "private key",
		SSHHostKey:    "host key",
	}}
	if _, err := p.sshSetup(); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"sshkey":           "private key",
		".ssh/known_hosts": "host key",
	} {
		path = filepath.Join(home, filepath.FromSlash(path))
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode = %v, want 0600", path, info.Mode().Perm())
		}
	}
}
