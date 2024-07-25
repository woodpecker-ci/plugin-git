package main

import (
	"os"
	"path/filepath"
	"testing"
)

// commits is a list of commits of different types (push, pull request, tag)
// to help us verify that this clone plugin can handle multiple commit types.
var commits = []struct {
	path      string
	clone     string
	event     string
	branch    string
	commit    string
	ref       string
	file      string
	data      string
	dataSize  int64
	recursive bool
	lfs       bool
}{
	// first commit
	{
		path:   "octocat/Hello-World",
		clone:  "https://github.com/octocat/Hello-World.git",
		event:  "push",
		branch: "master",
		commit: "553c2077f0edc3d5dc5d17262f6aa498e69d6f8e",
		ref:    "",
		file:   "README",
		data:   "Hello World!",
	},
	// head commit
	{
		path:   "octocat/Hello-World",
		clone:  "https://github.com/octocat/Hello-World.git",
		event:  "push",
		branch: "master",
		commit: "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
		ref:    "",
		file:   "README",
		data:   "Hello World!\n",
	},
	// pull request commit
	{
		path:   "octocat/Hello-World",
		clone:  "https://github.com/octocat/Hello-World.git",
		event:  "pull_request",
		branch: "master",
		commit: "762941318ee16e59dabbacb1b4049eec22f0d303",
		ref:    "",
		file:   "README",
		data:   "Hello World!\n",
	},
	// branch
	{
		path:   "octocat/Hello-World",
		clone:  "https://github.com/octocat/Hello-World.git",
		event:  "push",
		branch: "test",
		commit: "b3cbd5bbd7e81436d2eee04537ea2b4c0cad4cdf",
		ref:    "",
		file:   "CONTRIBUTING.md",
		data:   "## Contributing\n",
	},
	// tags
	{
		path:   "github/mime-types",
		clone:  "https://github.com/github/mime-types.git",
		event:  "tag",
		branch: "master",
		commit: "bf68d60215a167c935bc5976b7d06a7ffb290926",
		ref:    "",
		file:   ".gitignore",
		data:   "*.swp\n*~\n.rake_tasks~\nhtml\ndoc\npkg\npublish\ncoverage\n",
	},
	// submodules
	{
		path:      "test-assets/woodpecker-git-test-submodule",
		clone:     "https://github.com/test-assets/woodpecker-git-test-submodule.git",
		event:     "push",
		branch:    "main",
		commit:    "cc020eb6aaa601c13ca7b0d5db9d1ca694e7a003",
		ref:       "",
		file:      "Hello-World/README",
		data:      "Hello World!\n",
		recursive: true,
	},
	// checkout with ref only
	{
		path:  "octocat/Hello-World",
		clone: "https://github.com/octocat/Hello-World.git",
		event: "push",
		// commit: "a11fb45a696bf1d696fc9ab2c733f8f123aa4cf5",
		ref:  "pull/2403/head",
		file: "README",
		data: "Hello World!\n\nsomething is changed!\n",
	},
	// checkout with short SHA!
	{
		path:   "octocat/Hello-World",
		clone:  "https://github.com/octocat/Hello-World.git",
		event:  "pull_request",
		branch: "test",
		commit: "7629413",
		ref:    "",
		file:   "README",
		data:   "Hello World!\n",
	},
	// ### test lfs, please do not change order, otherwise TestCloneNonEmpty will fail ###
	// checkout with lfs skip
	{
		path:     "test-assets/woodpecker-git-test-lfs",
		clone:    "https://github.com/test-assets/woodpecker-git-test-lfs.git",
		event:    "push",
		commit:   "69d4dadb4c2899efb73c0095bb58a6454d133cef",
		ref:      "",
		file:     "4M.bin",
		dataSize: 132,
	},
	// checkout with lfs
	{
		path:     "test-assets/woodpecker-git-test-lfs",
		clone:    "https://github.com/test-assets/woodpecker-git-test-lfs.git",
		event:    "push",
		commit:   "69d4dadb4c2899efb73c0095bb58a6454d133cef",
		ref:      "",
		file:     "4M.bin",
		dataSize: 4194304,
		lfs:      true,
	},
}

// TestClone tests the ability to clone a specific commit into
// a fresh, empty directory every time.
func TestClone(t *testing.T) {
	for _, c := range commits {
		dir := setup()
		defer teardown(dir)

		plugin := Plugin{
			Repo: Repo{
				Clone: c.clone,
			},
			Pipeline: Pipeline{
				Path:   filepath.Join(dir, c.path),
				Commit: c.commit,
				Event:  c.event,
				Ref:    c.ref,
			},
			Config: Config{
				Recursive: c.recursive,
				Lfs:       c.lfs,
				Home:      "/tmp",
				Branch:    c.branch,
			},
		}

		if err := plugin.Exec(); err != nil {
			t.Errorf("Expected successful clone. Got error. %s.", err)
		}

		if c.data != "" {
			data := readFile(plugin.Pipeline.Path, c.file)
			if data != c.data {
				t.Errorf("Expected %s to contain [%s]. Got [%s].", c.file, c.data, data)
			}
		}

		if c.dataSize != 0 {
			size := getFileSize(plugin.Pipeline.Path, c.file)
			if size != c.dataSize {
				t.Errorf("Expected %s size to be [%d]. Got [%d].", c.file, c.dataSize, size)
			}
		}

	}
}

// TestCloneNonEmpty tests the ability to clone a specific commit into
// a non-empty directory. This is useful if the git workspace is cached
// and re-stored for every workflow.
func TestCloneNonEmpty(t *testing.T) {
	dir := setup()
	defer teardown(dir)

	for _, c := range commits {

		plugin := Plugin{
			Repo: Repo{
				Clone: c.clone,
			},
			Pipeline: Pipeline{
				Path:   filepath.Join(dir, c.path),
				Commit: c.commit,
				Event:  c.event,
				Ref:    c.ref,
			},
			Config: Config{
				Recursive: c.recursive,
				Lfs:       c.lfs,
				Home:      "/tmp",
				Branch:    c.branch,
			},
		}

		if err := plugin.Exec(); err != nil {
			t.Errorf("Expected successful clone. Got error. %s.", err)
		}

		if c.data != "" {
			data := readFile(plugin.Pipeline.Path, c.file)
			if data != c.data {
				t.Errorf("Expected %s to contain [%s]. Got [%s].", c.file, c.data, data)
				break
			}
		}

		if c.dataSize != 0 {
			size := getFileSize(plugin.Pipeline.Path, c.file)
			if size != c.dataSize {
				t.Errorf("Expected %s size to be [%d]. Got [%d].", c.file, c.dataSize, size)
			}
		}
	}
}

// TestFetch tests if the arguments to `git fetch` are constructed properly.
func TestFetch(t *testing.T) {
	testdata := []struct {
		ref   string
		tags  bool
		depth int
		exp   []string
	}{
		{
			"refs/heads/master",
			false,
			0,
			[]string{
				"git",
				"fetch",
				"--no-tags",
				"origin",
				"+refs/heads/master:",
			},
		},
		{
			"refs/heads/master",
			false,
			50,
			[]string{
				"git",
				"fetch",
				"--no-tags",
				"--depth=50",
				"origin",
				"+refs/heads/master:",
			},
		},
		{
			"refs/heads/master",
			true,
			100,
			[]string{
				"git",
				"fetch",
				"--tags",
				"--depth=100",
				"origin",
				"+refs/heads/master:",
			},
		},
	}
	for _, td := range testdata {
		c := fetch(td.ref, td.tags, td.depth, "")
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

// TestUpdateSubmodules tests if the arguments to `git submodule update`
// are constructed properly.
func TestUpdateSubmodules(t *testing.T) {
	testdata := []struct {
		partial bool
		exp     []string
	}{
		{
			false,
			[]string{
				"git",
				"submodule",
				"update",
				"--init",
				"--recursive",
			},
		},
		{
			true,
			[]string{
				"git",
				"submodule",
				"update",
				"--init",
				"--recursive",
				"--depth=1",
				"--recommend-shallow",
			},
		},
	}
	for _, td := range testdata {
		c := updateSubmodules(false, td.partial)
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

func TestCustomCertUrl(t *testing.T) {
	testdata := []struct {
		exp []string
	}{
		{
			[]string{
				"git",
				"config",
				"--global",
				"http.sslCAInfo",
				customCertTmpPath,
			},
		},
	}
	for _, td := range testdata {
		c := customCertHandler("http://example.com")
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

func TestCustomCertFile(t *testing.T) {
	testdata := []struct {
		exp []string
	}{
		{
			[]string{
				"git",
				"config",
				"--global",
				"http.sslCAInfo",
				"/etc/ssl/my-cert.pem",
			},
		},
	}

	for _, td := range testdata {
		c := customCertHandler("/etc/ssl/my-cert.pem")
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

func TestSwitchBranch(t *testing.T) {
	testdata := []struct {
		exp []string
	}{
		{
			[]string{
				"git",
				"switch",
				"-q",
				"test",
			},
		},
	}

	for _, td := range testdata {
		c := switchBranch("test")
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

// TestUpdateSubmodules tests if the arguments to `git submodule update`
// are constructed properly.
func TestUpdateSubmodulesRemote(t *testing.T) {
	testdata := []struct {
		exp []string
	}{
		{
			[]string{
				"git",
				"submodule",
				"update",
				"--init",
				"--recursive",
				"--remote",
			},
		},
		{
			[]string{
				"git",
				"submodule",
				"update",
				"--init",
				"--recursive",
				"--remote",
			},
		},
	}
	for _, td := range testdata {
		c := updateSubmodules(true, false)
		if len(c.Args) != len(td.exp) {
			t.Errorf("Expected: %s, got %s", td.exp, c.Args)
		}
		for i := range c.Args {
			if c.Args[i] != td.exp[i] {
				t.Errorf("Expected: %s, got %s", td.exp, c.Args)
			}
		}
	}
}

// helper function that will setup a temporary workspace.
// to which we can clone the repositroy
func setup() string {
	dir, _ := os.MkdirTemp("/tmp", "plugin_git_test_")
	os.Mkdir(dir, 0o777)
	return dir
}

// helper function to delete the temporary workspace.
func teardown(dir string) {
	os.RemoveAll(dir)
}

// helper function to read a file in the temporary worskapce.
func readFile(dir, file string) string {
	filename := filepath.Join(dir, file)
	data, _ := os.ReadFile(filename)
	return string(data)
}

func getFileSize(dir, file string) int64 {
	filename := filepath.Join(dir, file)
	fi, _ := os.Stat(filename)
	return fi.Size()
}
