package netrc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type NetRC struct {
	Machine,
	Login,
	Password string
}

func getFilePath() (string, error) {
	// get home
	homeDir := getHomeDir()

	// calc netrc file path
	file := filepath.Join(homeDir, ".netrc")
	if runtime.GOOS == "windows" {
		file = filepath.Join(homeDir, "_netrc")
	}

	stats, err := os.Stat(file)
	if err != nil {
		return "", nil
	}
	if !stats.Mode().IsRegular() {
		return "", fmt.Errorf("'%s' exist but is a %s", file, stats.Mode().Type().String())
	}
	return file, nil
}

// Delete delete the netrc if file exist
func Delete() (bool, error) {
	file, err := getFilePath()
	if err != nil || file == "" {
		return false, err
	}

	return true, os.Remove(file)
}

// Save save a netrc
func Save(n *NetRC) error {
	if n == nil {
		return nil
	}

	// get home
	homeDir := getHomeDir()

	// calc netrc file path
	file := filepath.Join(homeDir, ".netrc")
	if runtime.GOOS == "windows" {
		file = filepath.Join(homeDir, "_netrc")
	}

	content := fmt.Sprintf(`
machine %s
login %s
password %s
`,
		n.Machine,
		n.Login,
		n.Password,
	)

	return os.WriteFile(file, []byte(content), 0o600)
}

// Read return netrc if file or env var exist
func Read() (*NetRC, error) {
	// try to read from env var
	netRC := &NetRC{
		Machine:  os.Getenv("CI_NETRC_MACHINE"),
		Login:    os.Getenv("CI_NETRC_USERNAME"),
		Password: os.Getenv("CI_NETRC_PASSWORD"),
	}
	// if we get at least user and password from env we can return
	if netRC.Login != "" && netRC.Password != "" {
		return netRC, nil
	}

	file, err := getFilePath()
	if err != nil || file == "" {
		return nil, err
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while reading file '%s': %w", file, err)
	}

	for _, v := range strings.Split(string(raw), "\n") {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "machine") {
			netRC.Machine = strings.TrimSpace(strings.TrimPrefix(v, "machine"))
		}
		if strings.HasPrefix(v, "login") {
			netRC.Login = strings.TrimSpace(strings.TrimPrefix(v, "login"))
		}
		if strings.HasPrefix(v, "password") {
			netRC.Password = strings.TrimSpace(strings.TrimPrefix(v, "password"))
		}
	}

	return nil, nil
}

func getHomeDir() string {
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		return homeDir
	}
	if homeDir, _ := os.UserHomeDir(); homeDir != "" {
		return homeDir
	}
	pwd, _ := os.Getwd()
	return pwd
}
