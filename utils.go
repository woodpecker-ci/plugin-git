package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// logGroupMarker starts a collapsible output group in the Woodpecker web UI
// when printed at the beginning of a line followed by a title.
const logGroupMarker = "▶  "

// trace writes the command to stdout as an output group heading, so the
// web UI groups the command's output under it.
func trace(cmd *exec.Cmd) {
	fmt.Printf("%s%s\n", logGroupMarker, strings.Join(cmd.Args, " "))
}

// pathExists returns whether the given file or directory exists or not
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// helper function returns true if directory dir is empty.
func isDirEmpty(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return true
	}
	defer f.Close()

	_, err = f.Readdir(1)
	return err == io.EOF
}

// helper function to write a netrc file.
func writeNetrc(home, machine, login, password string) error {
	if machine == "" || (login == "" && password == "") {
		return nil
	}
	out := fmt.Sprintf(
		netrcFile,
		machine,
		login,
		password,
	)

	path := filepath.Join(home, ".netrc")
	return os.WriteFile(path, []byte(out), 0o600)
}

const netrcFile = `
machine %s
login %s
password %s
`
