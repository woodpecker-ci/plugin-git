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
	Repo    Repo
	Build   Build
	Netrc   Netrc
	Config  Config
	Backoff Backoff
}

const customCertTmpPath = "/tmp/customCert.pem"

func (p Plugin) Exec() error {
	if p.Build.Path != "" {
		err := os.MkdirAll(p.Build.Path, 0o777)
		if err != nil {
			return err
		}
	}

	err := writeNetrc(p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
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

	if isDirEmpty(filepath.Join(p.Build.Path, ".git")) {
		cmds = append(cmds, initGit())
		cmds = append(cmds, remote(p.Repo.Clone))
	}

	switch {
	case isPullRequest(p.Build.Event) || isTag(p.Build.Event, p.Build.Ref):
		cmds = append(cmds, fetch(p.Build.Ref, p.Config.Tags, p.Config.Depth, !p.Config.Lfs))
		cmds = append(cmds, checkoutHead())
	default:
		cmds = append(cmds, fetch(p.Build.Ref, p.Config.Tags, p.Config.Depth, !p.Config.Lfs))
		cmds = append(cmds, checkoutSha(p.Build.Commit))
	}

	for name, submoduleUrl := range p.Config.Submodules {
		cmds = append(cmds, remapSubmodule(name, submoduleUrl))
	}

	if p.Config.Recursive {
		cmds = append(cmds, updateSubmodules(p.Config.SubmoduleRemote))
	}

	for _, cmd := range cmds {
		buf := new(bytes.Buffer)
		cmd.Dir = p.Build.Path
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

// Creates an empty git repository.
func initGit() *exec.Cmd {
	return exec.Command(
		"git",
		"init",
	)
}

// Sets the remote origin for the repository.
func remote(remote string) *exec.Cmd {
	return exec.Command(
		"git",
		"remote",
		"add",
		"origin",
		remote,
	)
}

// Checkout executes a git checkout command.
func checkoutHead() *exec.Cmd {
	return exec.Command(
		"git",
		"checkout",
		"-qf",
		"FETCH_HEAD",
	)
}

// Checkout executes a git checkout command.
func checkoutSha(commit string) *exec.Cmd {
	return exec.Command(
		"git",
		"reset",
		"--hard",
		"-q",
		commit,
	)
}

// fetch retuns git command that fetches from origin. If tags is true
// then tags will be fetched.
func fetch(ref string, tags bool, depth int, skipLfs bool) *exec.Cmd {
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
	cmd.Args = append(cmd.Args, "origin")
	cmd.Args = append(cmd.Args, fmt.Sprintf("+%s:", ref))
	if skipLfs {
		// The GIT_LFS_SKIP_SMUDGE env var prevents git-lfs from retrieving any
		// LFS files.
		cmd.Env = append(cmd.Env, "GIT_LFS_SKIP_SMUDGE=1")
	}

	return cmd
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

	return cmd
}

// skipVerify returns a git command that, when executed configures git to skip
// ssl verification. This should may be used with self-signed certificates.
func skipVerify() *exec.Cmd {
	return exec.Command(
		"git",
		"config",
		"--global",
		"http.sslVerify",
		"false",
	)
}

func setCustomCert(path string) *exec.Cmd {
	return exec.Command(
		"git",
		"config",
		"--global",
		"http.sslCAInfo",
		path,
	)
}

// remapSubmodule returns a git command that, when executed configures git to
// remap submodule urls.
func remapSubmodule(name, url string) *exec.Cmd {
	name = fmt.Sprintf("submodule.%s.url", name)
	return exec.Command(
		"git",
		"config",
		"--global",
		name,
		url,
	)
}
