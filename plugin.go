package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

var customCertTmpPath = filepath.Join(os.TempDir(), "customCert.pem")

var defaultEnvVars = []string{
	// do not set GIT_TERMINAL_PROMPT=0, otherwise git won't load credentials from ".netrc"
	"GIT_LFS_SKIP_SMUDGE=1", // prevents git-lfs from retrieving any LFS files
}

func (p Plugin) Exec() error {
	p.normalizeConfig()

	if err := p.prepareWorkspace(); err != nil {
		return err
	}
	baseEnv := append(os.Environ(), "HOME="+p.Config.Home)

	cmds, err := p.commands(baseEnv)
	if err != nil {
		return err
	}
	return p.run(cmds, baseEnv)
}

// normalizeConfig fills in derived settings and resolves conflicting ones.
func (p *Plugin) normalizeConfig() {
	p.Config.Sparse = normalizeSparseEntries(p.Config.Sparse)

	// autodetect object-format if not set
	if p.Repo.ObjectFormat == "" {
		p.Repo.ObjectFormat = "sha1"
		if len(p.Pipeline.Commit) == 64 {
			p.Repo.ObjectFormat = "sha256"
		}
	}

	// depth should be set to 0 if merge pull request is enabled
	if p.Config.MergePullRequest && p.Config.Depth != 0 {
		p.Config.Depth = 0
	}
}

// prepareWorkspace creates the workspace and sets up the home directory used
// by git.
func (p Plugin) prepareWorkspace() error {
	// set umask (default 0) so cloned files are
	// accessible from non-root containers
	umask(p.Config.Umask)

	if p.Pipeline.Path != "" {
		err := os.MkdirAll(p.Pipeline.Path, 0o777)
		if err != nil {
			return err
		}
	}

	// make sure pipeline workspace is always writable for anyone
	os.Chmod(p.Pipeline.Path, 0o777)

	err := writeNetrc(p.Config.Home, p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
		return err
	}

	return setHome(p.Config.Home)
}

// commands returns all git commands needed to clone the repository, in the
// order they have to be executed.
func (p Plugin) commands(baseEnv []string) ([]*exec.Cmd, error) {
	sparseEnabled := len(p.Config.Sparse) > 0
	// in sparse mode LFS files are smudged while they are checked out, so only
	// the LFS objects inside the sparse cone are downloaded
	sparseLFS := sparseEnabled && p.Config.Lfs

	newRepository := isDirEmpty(filepath.Join(p.Pipeline.Path, ".git"))

	cmds := p.sslCommands()

	if newRepository {
		initCmds, err := p.initCommands()
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, initCmds...)
	}

	fetchCmd, checkoutCmds, err := p.fetchCommands()
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, fetchCmd)

	sparseCmds, err := p.sparseCommands(newRepository, baseEnv)
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, sparseCmds...)

	for _, cmd := range checkoutCmds {
		cmds = append(cmds, withLFSSmudge(cmd, sparseLFS))
	}

	submoduleCmds, err := p.submoduleCommands()
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, submoduleCmds...)

	pullRequestCmds, err := p.pullRequestCommands(sparseLFS)
	if err != nil {
		return nil, err
	}
	cmds = append(cmds, pullRequestCmds...)

	if p.Config.Lfs && !sparseEnabled {
		cmds = append(cmds,
			fetchLFS(),
			checkoutLFS())
	}

	return cmds, nil
}

// sslCommands configures how the TLS certificate of the remote is verified.
func (p Plugin) sslCommands() []*exec.Cmd {
	if p.Config.SkipVerify {
		return []*exec.Cmd{skipVerify()}
	}
	if p.Config.CustomCert != "" {
		if certCmd := customCertHandler(p.Config.CustomCert); certCmd != nil {
			return []*exec.Cmd{certCmd}
		}
	}
	return nil
}

// initCommands initializes a new repository and configures its remote.
func (p Plugin) initCommands() ([]*exec.Cmd, error) {
	cmds := []*exec.Cmd{
		initGit(p.Config.Branch, p.Repo.ObjectFormat),
		safeDirectory(p.Config.SafeDirectory),
	}
	if !p.Config.UseSSH {
		return append(cmds, remote(p.Repo.Clone)), nil
	}

	// If env var PLUGIN_USE_SSH is set to true, use SSH instead of HTTPS
	cmds = append(cmds, remote(p.Repo.CloneSSH))
	sshCmd, err := p.sshSetup()
	if err != nil {
		return nil, err
	}
	if sshCmd != nil {
		cmds = append(cmds, sshCmd)
	}
	return cmds, nil
}

// sshSetup writes the configured SSH key and host key to the home directory
// and returns the command configuring git to use them, if any.
func (p Plugin) sshSetup() (*exec.Cmd, error) {
	sshKey := p.Config.SSHKey
	if p.Config.SSHKeyPrivate != "" {
		if sshKey == "" {
			sshKey = filepath.Join(p.Config.Home, "/sshkey")
		}
		if err := os.WriteFile(sshKey, []byte(p.Config.SSHKeyPrivate), 0o600); err != nil {
			return nil, fmt.Errorf("could not write private SSH key: %v", err)
		}
	}

	knownHosts := ""
	if p.Config.SSHHostKey != "" {
		if err := os.Mkdir(filepath.Join(p.Config.Home, ".ssh"), 0o700); err != nil {
			return nil, fmt.Errorf("could not create SSH directory: %v", err)
		}
		knownHosts = filepath.Join(p.Config.Home, ".ssh/known_hosts")
		if err := os.WriteFile(knownHosts, []byte(p.Config.SSHHostKey), 0o600); err != nil {
			return nil, fmt.Errorf("could not write SSH host key: %v", err)
		}
	}

	if sshKey == "" {
		return nil, nil
	}
	// If env var PLUGIN_SSH_KEY is set, use it as the SSH key
	return sshKeyHandler(sshKey, knownHosts), nil
}

// fetchCommands returns the command fetching the commit to build and the
// commands checking it out.
func (p Plugin) fetchCommands() (*exec.Cmd, []*exec.Cmd, error) {
	if p.Pipeline.Ref != "" {
		// fetch and checkout by ref
		fmt.Println("using head checkout")
		return fetch(p.Pipeline.Ref, p.Config.Tags, p.Config.Depth, p.Config.filter),
			[]*exec.Cmd{checkoutHead()}, nil
	}

	if len(p.Pipeline.Commit) != 40 && len(p.Pipeline.Commit) != 64 {
		// fetch requires full SHA1 (40 chars) or SHA256 (64 chars) commits (unambiguous reference)
		// for short SHA1 or SHA256, fetch and switch the branch before commit reset
		if p.Config.Branch == "" {
			return nil, nil, fmt.Errorf("short commit SHA1 checkout requires a branch")
		}
		return fetch(p.Config.Branch, p.Config.Tags, p.Config.Depth, p.Config.filter),
			[]*exec.Cmd{switchBranch(p.Config.Branch), checkoutSha(p.Pipeline.Commit)}, nil
	}

	// fetch and checkout by commit sha
	return fetch(p.Pipeline.Commit, p.Config.Tags, p.Config.Depth, p.Config.filter),
		[]*exec.Cmd{checkoutSha(p.Pipeline.Commit)}, nil
}

// submoduleCommands remaps submodule urls and updates submodules.
func (p Plugin) submoduleCommands() ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd
	if p.Config.Submodules != "" {
		var submoduleOverrides map[string]string
		err := json.Unmarshal([]byte(p.Config.Submodules), &submoduleOverrides)
		if err != nil {
			return nil, fmt.Errorf("could not parse submodule_override map: %v", err)
		}
		for name, submoduleUrl := range submoduleOverrides {
			cmds = append(cmds, remapSubmodule(name, submoduleUrl))
		}
	}

	if p.Config.Recursive {
		cmds = append(cmds, updateSubmodules(p.Config.SubmoduleRemote, p.Config.SubmodulePartial))
	}
	return cmds, nil
}

// pullRequestCommands merges or fetches the target branch of pull requests.
func (p Plugin) pullRequestCommands(lfsSmudge bool) ([]*exec.Cmd, error) {
	if p.Config.Event != "pull_request" {
		return nil, nil
	}

	if p.Config.MergePullRequest {
		return []*exec.Cmd{
			setUserName(p.Config.GitUserName),
			setUserEmail(p.Config.GitUserEmail),
			fetchBranch(p.Config.TargetBranch),
			withLFSSmudge(mergeBranch(p.Config.TargetBranch), lfsSmudge),
		}, nil
	}

	if p.Config.FetchTargetBranch {
		if p.Config.TargetBranch == "" {
			return nil, fmt.Errorf("fetch_target_branch is enabled but target_branch is not set")
		}
		return []*exec.Cmd{fetchBranch(p.Config.TargetBranch)}, nil
	}
	return nil, nil
}

// run executes the commands in the workspace, retrying them when the remote
// ref is not available yet.
func (p Plugin) run(cmds []*exec.Cmd, baseEnv []string) error {
	for _, cmd := range cmds {
		cmd.Env = slices.Concat(baseEnv, cmd.Env)
		cmd.Dir = p.Pipeline.Path
		buf := new(bytes.Buffer)
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
	for range retries {
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

func normalizeSparseEntries(entries []string) []string {
	if len(entries) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry != "" {
			normalized = append(normalized, entry)
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

// sparseCommands returns the commands that configure the sparse checkout of
// the workspace, or that restore a full checkout of a cached sparse workspace.
func (p Plugin) sparseCommands(newRepository bool, env []string) ([]*exec.Cmd, error) {
	if len(p.Config.Sparse) == 0 {
		if newRepository {
			return nil, nil
		}
		enabled, err := isSparseCheckoutEnabled(p.Pipeline.Path, slices.Concat(env, defaultEnvVars))
		if err != nil || !enabled {
			return nil, err
		}
		return []*exec.Cmd{disableSparseCheckout()}, nil
	}

	var cmds []*exec.Cmd
	if p.Config.Lfs {
		cmds = append(cmds, installLFS())
	}
	fmt.Printf("configuring sparse checkout (%d directories)\n", len(p.Config.Sparse))
	return append(cmds, withLFSSmudge(sparseCheckout(p.Config.Sparse), p.Config.Lfs)), nil
}

// withLFSSmudge lets git-lfs download and materialize the LFS files written
// by cmd, instead of leaving pointer files behind.
func withLFSSmudge(cmd *exec.Cmd, enabled bool) *exec.Cmd {
	if enabled {
		cmd.Env = slices.DeleteFunc(cmd.Env, func(env string) bool {
			return slices.Contains(defaultEnvVars, env)
		})
	}
	return cmd
}

func sparseCheckout(paths []string) *exec.Cmd {
	args := append([]string{"sparse-checkout", "set", "--cone", "--"}, paths...)
	return appendEnv(exec.Command("git", args...), defaultEnvVars...)
}

func disableSparseCheckout() *exec.Cmd {
	return appendEnv(exec.Command("git", "sparse-checkout", "disable"), defaultEnvVars...)
}

func installLFS() *exec.Cmd {
	return appendEnv(exec.Command("git", "lfs", "install", "--local"), defaultEnvVars...)
}

func isSparseCheckoutEnabled(dir string, env []string) (bool, error) {
	cmd := exec.Command("git", "config", "--bool", "--get", "core.sparseCheckout")
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("could not inspect cached sparse checkout state: %w", err)
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

func appendEnv(cmd *exec.Cmd, env ...string) *exec.Cmd {
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

// Creates an empty git repository.
func initGit(branch, objectFormat string) *exec.Cmd {
	if branch == "" {
		return appendEnv(exec.Command("git", "init", "--object-format", objectFormat), defaultEnvVars...)
	}
	return appendEnv(exec.Command("git", "init", "--object-format", objectFormat, "-b", branch), defaultEnvVars...)
}

func safeDirectory(safeDirectory string) *exec.Cmd {
	return appendEnv(exec.Command("git", "config", "--global", "--replace-all", "safe.directory", safeDirectory), defaultEnvVars...)
}

func setUserName(userName string) *exec.Cmd {
	return appendEnv(exec.Command("git", "config", "--global", "user.name", userName), defaultEnvVars...)
}

func setUserEmail(userEmail string) *exec.Cmd {
	return appendEnv(exec.Command("git", "config", "--global", "user.email", userEmail), defaultEnvVars...)
}

// Use custom SSH Key thanks to core.sshCommand
func sshKeyHandler(sshKey string, knownHosts string) *exec.Cmd {
	hostKeyArg := " -o \"StrictHostKeyChecking no\""
	if knownHosts != "" {
		hostKeyArg = " -o UserKnownHostsFile=" + knownHosts
	}
	return appendEnv(exec.Command("git", "config", "--global", "core.sshCommand", "ssh -i "+sshKey+hostKeyArg), defaultEnvVars...)
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

// Fetch a branch
func fetchBranch(branch string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"fetch",
		"origin",
		"--",
		branch,
	), defaultEnvVars...)
}

// Merge a branch
func mergeBranch(branch string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"merge",
		"origin/"+branch,
	), defaultEnvVars...)
}

// Switch executes a git switch command.
func switchBranch(branch string) *exec.Cmd {
	return appendEnv(exec.Command(
		"git",
		"switch",
		"-q",
		branch,
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
func updateSubmodules(remote, partial bool) *exec.Cmd {
	args := []string{"submodule", "update", "--init", "--recursive"}
	if partial {
		args = append(args, "--depth=1", "--recommend-shallow")
	}
	cmd := exec.Command("git", args...)

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
	return nil
}
