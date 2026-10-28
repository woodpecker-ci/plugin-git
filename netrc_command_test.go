package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestNetrcCommandHome(t *testing.T) {
	if os.Getenv("PLUGIN_GIT_TEST_NETRC_COMMAND") == "1" {
		os.Args = os.Args[slices.Index(os.Args, "--")+1:]
		main()
		return
	}
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
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestNetrcCommandHome$", "--"}, args...)...)
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
