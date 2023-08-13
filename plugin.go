package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Plugin struct {
	Repo     Repo
	Pipeline Pipeline
	Netrc    Netrc
	Config   Config
	Backoff  Backoff
}

const customCertTmpPath = "/tmp/customCert.pem"

var defaultEnvVars = []string{
	// do not set GIT_TERMINAL_PROMPT=0, otherwise git won't load credentials from ".netrc"
	"GIT_LFS_SKIP_SMUDGE=1", // prevents git-lfs from retrieving any LFS files
}

func (p Plugin) Exec() error {
	if p.Pipeline.Path != "" {
		err := os.MkdirAll(p.Pipeline.Path, 0o777)
		if err != nil {
			return err
		}
	}

	err := writeNetrc(p.Config.Home, p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
		return err
	}

	// set vars from exec environment
	defaultEnvVars = append(os.Environ(), defaultEnvVars...)

	// alter home var for all commands exec afterwards
	if err := setHome(p.Config.Home); err != nil {
		return err
	}

	var cmds []*exec.Cmd

	if p.Config.SkipVerify {
		cmds = append(cmds, skipVerify())
	} else if p.Config.CustomCert != "" {
		certCmd := customCertHandler(p.Config.CustomCert)
		if certCmd != nil {
			cmds = append(cmds, certCmd)
		}
	}

	if isDirEmpty(filepath.Join(p.Pipeline.Path, ".git")) {
		cmds = append(cmds, initGit(p.Config.Branch))
		cmds = append(cmds, safeDirectory(p.Config.SafeDirectory))
		cmds = append(cmds, remote(p.Repo.Clone))
	}

	// fetch ref in any case
	cmds = append(cmds, fetch(p.Pipeline.Ref, p.Config.Tags, p.Config.Depth, p.Config.filter))

	if p.Pipeline.Commit == "" {
		// checkout by fetched ref
		fmt.Println("no commit information: using head checkout")
		cmds = append(cmds, checkoutHead())
	} else {
		// checkout by commit sha
		cmds = append(cmds, checkoutSha(p.Pipeline.Commit))
	}

	for name, submoduleUrl := range p.Config.Submodules {
		cmds = append(cmds, remapSubmodule(name, submoduleUrl))
	}

	if p.Config.Recursive {
		cmds = append(cmds, updateSubmodules(p.Config.SubmoduleRemote))
	}

	if p.Config.Lfs {
		cmds = append(cmds,
			fetchLFS(),
			checkoutLFS())
	}

	for _, cmd := range cmds {
		buf := new(bytes.Buffer)
		cmd.Dir = p.Pipeline.Path
		cmd.Stdout = io.MultiWriter(os.Stdout, buf)
		cmd.Stderr = io.MultiWriter(os.Stderr, buf)
		trace(cmd)
		err := cmd.Run()
		switch {
		case err != nil && shouldRetry(buf.String()):
			err = retryExec(cmd, p.Backoff.Duration, p.Backoff.Attempts)
			if err != nil {
				return err
			}
		case err != nil:
			return err
		}
	}

	return nil
}

func customCertHandler(certPath string) *exec.Cmd {
	if IsUrl(certPath) {
		if downloadCert(certPath) {
			return setCustomCert(customCertTmpPath)
		} else {
			fmt.Printf("Failed to download custom ssl cert. Ignoring...\n")
			return nil
		}
	}
	return setCustomCert(certPath)
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func downloadCert(url string) (retStatus bool) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to download %s\n", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			retStatus = false
		}
	}(resp.Body)

	out, err := os.Create(customCertTmpPath)
	if err != nil {
		fmt.Printf("Failed to create file %s\n", customCertTmpPath)
		return false
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			retStatus = false
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to copy cert to %s\n", customCertTmpPath)
		return false
	}
	return true
}

// shouldRetry returns true if the command should be re-executed. Currently
// this only returns true if the remote ref does not exist.
func shouldRetry(s string) bool {
	return strings.Contains(s, "find remote ref")
}

// retryExec is a helper function that retries a command.
func retryExec(cmd *exec.Cmd, backoff time.Duration, retries int) (err error) {
	for i := 0; i < retries; i++ {
		// signal intent to retry
		fmt.Printf("retry in %v\n", backoff)

		// wait 5 seconds before retry
		<-time.After(backoff)

		// copy the original command
		retry := exec.Command(cmd.Args[0], cmd.Args[1:]...)
		retry.Dir = cmd.Dir
		retry.Env = cmd.Env
		retry.Stdout = os.Stdout
		retry.Stderr = os.Stderr
		trace(retry)
		err = retry.Run()
		if err == nil {
			return
		}
	}
	return
}

func appendEnv(cmd *exec.Cmd, env ...string) *exec.Cmd {
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

// Creates an empty git repository.
func initGit(branch string) *exec.Cmd {
	if branch == "" {
		return appendEnv(exec.Command("git", "init"), defaultEnvVars...)
	}
	return appendEnv(exec.Command("git", "init", "-b", branch), defaultEnvVars...)
}

func safeDirectory(safeDirectory string) *exec.Cmd {
	return appendEnv(exec.Command("git", "config", "--global", "safe.directory", safeDirectory), defaultEnvVars...)
}

// Sets the remote origin for the repository.
func remote(remote string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"remote",
		"add",
		"origin",
		remote,
	), defaultEnvVars...)
}

// Checkout executes a git checkout command.
func checkoutHead() *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"checkout",
		"-qf",
		"FETCH_HEAD",
	), defaultEnvVars...)
}

// Checkout executes a git checkout command.
func checkoutSha(commit string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"reset",
		"--hard",
		"-q",
		commit,
	), defaultEnvVars...)
}

func fetchLFS() *exec.Cmd {
	return appendEnv(exec.Command(
		"git", "lfs",
		"fetch",
	), defaultEnvVars...)
}

func checkoutLFS() *exec.Cmd {
	return appendEnv(exec.Command(
		"git", "lfs",
		"checkout",
	), defaultEnvVars...)
}

// fetch returns git command that fetches from origin. If tags is true
// then tags will be fetched.
func fetch(ref string, tags bool, depth int, filter string) *exec.Cmd {
	tags_option := "--no-tags"
	if tags {
		tags_option = "--tags"
	}
	cmd := exec.Command(
		"git",
		"fetch",
		tags_option,
	)
	if depth != 0 {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--depth=%d", depth))
	}
	if filter != "" {
		cmd.Args = append(cmd.Args, "--filter="+filter)
	}
	cmd.Args = append(cmd.Args, "origin")
	cmd.Args = append(cmd.Args, fmt.Sprintf("+%s:", ref))

	return appendEnv(cmd, defaultEnvVars...)
}

// updateSubmodules recursively initializes and updates submodules.
func updateSubmodules(remote bool) *exec.Cmd {
	cmd := exec.Command(
		"git",
		"submodule",
		"update",
		"--init",
		"--recursive",
	)

	if remote {
		cmd.Args = append(cmd.Args, "--remote")
	}

	return appendEnv(cmd, defaultEnvVars...)
}

// skipVerify returns a git command that, when executed configures git to skip
// ssl verification. This should may be used with self-signed certificates.
func skipVerify() *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"config",
		"--global",
		"http.sslVerify",
		"false",
	), defaultEnvVars...)
}

func setCustomCert(path string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"config",
		"--global",
		"http.sslCAInfo",
		path,
	), defaultEnvVars...)
}

// remapSubmodule returns a git command that, when executed configures git to
// remap submodule urls.
func remapSubmodule(name, url string) *exec.Cmd {
	name = fmt.Sprintf("submodule.%s.url", name)
	return appendEnv(exec.Command(
		"git",
		"config",
		"--global",
		name,
		url,
	), defaultEnvVars...)
}

func setHome(home string) error {
	// make sure home dir exist and is set
	homeExist, err := pathExists(home)
	if err != nil {
		return err
	}
	if !homeExist {
		return fmt.Errorf("home directory '%s' do not exist", home)
	}
	defaultEnvVars = append(defaultEnvVars, "HOME="+home)

	return nil
}
