package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	switch os.Getenv("PLUGIN_GIT_TEST_NETRC_COMMAND") {
	case "1":
		main()
		os.Exit(0)
	case "configure":
		command := setNetRCHelper()
		command.Env = append(os.Environ(), command.Env...)
		if err := command.Run(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNetrcCommandHome(t *testing.T) {
	for _, name := range []string{"global flag", "flag after command", "environment", "flag overrides environment"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			other := t.TempDir()
			t.Setenv("HOME", other)
			t.Setenv("PLUGIN_HOME", "")
			filename := ".netrc"
			if runtime.GOOS == "windows" {
				filename = "_netrc"
			}
			for _, directory := range []string{home, other} {
				if err := os.WriteFile(filepath.Join(directory, filename), []byte("test credentials"), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			args := []string{"plugin-git", "--home", home, "netrc", "erase"}
			switch name {
			case "flag after command":
				args = []string{"plugin-git", "netrc", "--home", home, "erase"}
			case "environment":
				t.Setenv("PLUGIN_HOME", home)
				args = []string{"plugin-git", "netrc", "erase"}
			case "flag overrides environment":
				t.Setenv("PLUGIN_HOME", other)
			}
			command := exec.Command(os.Args[0], args[1:]...)
			command.Env = append(os.Environ(), "PLUGIN_GIT_TEST_NETRC_COMMAND=1")
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("netrc erase failed: %v\n%s", err, output)
			}
			if _, err := os.Stat(filepath.Join(home, filename)); !os.IsNotExist(err) {
				t.Errorf("selected home netrc was not erased: %v", err)
			}
			if _, err := os.Stat(filepath.Join(other, filename)); err != nil {
				t.Errorf("unselected home netrc was changed: %v", err)
			}
		})
	}
}

func TestNetrcHelperExecutablePath(t *testing.T) {
	for _, name := range []string{"plugin-git.exe", "plugin git's $tool.exe"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			t.Setenv("GIT_CONFIG_COUNT", "0")
			t.Setenv("GIT_CONFIG_PARAMETERS", "")
			t.Setenv("GIT_TERMINAL_PROMPT", "0")
			t.Setenv("CI_NETRC_MACHINE", "example.invalid")
			t.Setenv("CI_NETRC_USERNAME", "test-user")
			t.Setenv("CI_NETRC_PASSWORD", "test-password")
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			binary, err := os.ReadFile(executable)
			if err != nil {
				t.Fatal(err)
			}
			executable = filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(executable, binary, 0o755); err != nil {
				t.Fatal(err)
			}
			configure := exec.Command(executable)
			configure.Env = append(os.Environ(), "PLUGIN_GIT_TEST_NETRC_COMMAND=configure")
			if output, err := configure.CombinedOutput(); err != nil {
				t.Fatalf("configure helper: %v\n%s", err, output)
			}
			fill := exec.Command("git", "credential", "fill")
			fill.Env = append(os.Environ(), "PLUGIN_GIT_TEST_NETRC_COMMAND=1")
			fill.Stdin = strings.NewReader("protocol=https\nhost=example.invalid\n\n")
			output, err := fill.CombinedOutput()
			if err != nil {
				t.Fatalf("credential fill: %v\n%s", err, output)
			}
			for _, field := range []string{"username=test-user\n", "password=test-password\n"} {
				if !strings.Contains(string(output), field) {
					t.Errorf("missing credential field %q: %s", field, output)
				}
			}
		})
	}
}
