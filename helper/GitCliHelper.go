package helper

import (
	"fmt"
	"github.com/devtron-labs/ci-runner/util"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepoManager interface {
	Fetch(gitContext GitContext, rootDir string) (response, errMsg string, err error)
	Checkout(gitContext GitContext, rootDir string, checkout string) (response, errMsg string, err error)
	RunCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error)
	RunCommand(cmd *exec.Cmd) (response, errMsg string, err error)
	runCommandForSuppliedNullifiedEnv(cmd *exec.Cmd, setHomeEnvToNull bool) (response, errMsg string, err error)
	Init(rootDir string, remoteUrl string, isBare bool) error
	Clone(gitContext GitContext, prj CiProjectDetails) (response, errMsg string, err error)
	Merge(rootDir string, commit string) (response, errMsg string, err error)
	RecursiveFetchSubmodules(rootDir string) (response, errMsg string, error error)
	UpdateCredentialHelper(rootDir string) (response, errMsg string, error error)
	UnsetCredentialHelper(rootDir string) (response, errMsg string, error error)
	GitCheckout(gitContext GitContext, checkoutPath string, targetCheckout string, authMode AuthMode, fetchSubmodules bool, gitRepository string) (errMsg string, error error)
}

type RepoManagerImpl struct {
}

func NewRepoManager() *RepoManagerImpl {
	return &RepoManagerImpl{}
}

const GIT_AKS_PASS = "/git-ask-pass.sh"

func (impl *RepoManagerImpl) Fetch(gitContext GitContext, rootDir string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.RunCommandWithCred(cmd, gitContext.Auth.Username, gitContext.Auth.Password)
	log.Println(util.DEVTRON, "fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, "", nil
}

func (impl *RepoManagerImpl) Checkout(gitContext GitContext, rootDir string, checkout string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git checkout ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "checkout", checkout, "--force")
	output, errMsg, err := impl.RunCommandWithCred(cmd, gitContext.Auth.Username, gitContext.Auth.Password)
	log.Println(util.DEVTRON, "checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, "", nil
}

func (impl *RepoManagerImpl) RunCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_AKS_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName), // ignored
		fmt.Sprintf("GIT_PASSWORD=%s", password), // this value is used
	)
	return impl.RunCommand(cmd)
}

func (impl *RepoManagerImpl) RunCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	return impl.runCommandForSuppliedNullifiedEnv(cmd, true)
}

func (impl *RepoManagerImpl) runCommandForSuppliedNullifiedEnv(cmd *exec.Cmd, setHomeEnvToNull bool) (response, errMsg string, err error) {
	if setHomeEnvToNull {
		cmd.Env = append(cmd.Env, "HOME=/dev/null")
	}
	// https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
	// in CombinedOutput, both stdOut and stdError are returned in single output
	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)
	output = strings.Replace(output, "\n", "", -1)
	output = strings.TrimSpace(output)
	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", output, err
		}
		errOutput := string(exErr.Stderr)
		return "", fmt.Sprintf("%s\n%s", output, errOutput), err
	}
	return output, "", nil
}

func (impl *RepoManagerImpl) Init(rootDir string, remoteUrl string, isBare bool) error {

	//-----------------

	err := os.MkdirAll(rootDir, 0755)
	if err != nil {
		return err
	}
	repo, err := git.PlainInit(rootDir, isBare)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remoteUrl},
	})
	return err
}

func (impl *RepoManagerImpl) Clone(gitContext GitContext, prj CiProjectDetails) (response, errMsg string, err error) {
	rootDir := filepath.Join(util.WORKINGDIR, prj.CheckoutPath)
	remoteUrl := prj.GitRepository
	err = impl.Init(rootDir, remoteUrl, false)
	if err != nil {
		return "", "", err
	}

	response, errMsg, err = impl.Fetch(gitContext, rootDir)
	return response, errMsg, err
}

// setting user.name and user.email as for non-fast-forward merge, git ask for user.name and email
func (impl *RepoManagerImpl) Merge(rootDir string, commit string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git merge ", "location", rootDir)
	command := "cd " + rootDir + " && git config user.email git@devtron.com && git config user.name Devtron && git merge " + commit + " --no-commit"
	cmd := exec.Command("/bin/sh", "-c", command)
	output, errMsg, err := impl.RunCommand(cmd)
	log.Println(util.DEVTRON, "merge output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *RepoManagerImpl) RecursiveFetchSubmodules(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git recursive fetch submodules ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "submodule", "update", "--init", "--recursive")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "recursive fetch submodules output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *RepoManagerImpl) UpdateCredentialHelper(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git credential helper store ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "config", "--global", "credential.helper", "store")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "git credential helper store output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *RepoManagerImpl) UnsetCredentialHelper(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git credential helper unset ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "config", "--global", "--unset", "credential.helper")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "git credential helper unset output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *RepoManagerImpl) GitCheckout(gitContext GitContext, checkoutPath string, targetCheckout string, authMode AuthMode, fetchSubmodules bool, gitRepository string) (errMsg string, error error) {

	rootDir := filepath.Join(util.WORKINGDIR, checkoutPath)

	// checkout target hash
	_, eMsg, cErr := impl.Checkout(gitContext, rootDir, targetCheckout)
	if cErr != nil {
		return eMsg, cErr
	}

	log.Println(util.DEVTRON, " fetchSubmodules ", fetchSubmodules, " authMode ", authMode)

	if fetchSubmodules {
		httpsAuth := (authMode == AUTH_MODE_USERNAME_PASSWORD) || (authMode == AUTH_MODE_ACCESS_TOKEN)
		if httpsAuth {
			// first remove protocol
			modifiedUrl := strings.ReplaceAll(gitRepository, "https://", "")
			// for bitbucket - if git repo url is started with username, then we need to remove username
			if strings.Contains(modifiedUrl, "bitbucket.org") && !strings.HasPrefix(modifiedUrl, "bitbucket.org") {
				modifiedUrl = modifiedUrl[strings.Index(modifiedUrl, "bitbucket.org"):]
			}
			// build url
			modifiedUrl = "https://" + gitContext.Auth.Username + ":" + gitContext.Auth.Password + "@" + modifiedUrl

			_, errMsg, cErr = impl.UpdateCredentialHelper(rootDir)
			if cErr != nil {
				return errMsg, cErr
			}

			cErr = util.CreateGitCredentialFileAndWriteData(modifiedUrl)
			if cErr != nil {
				return "Error in creating git credential file", cErr
			}

		}

		_, errMsg, cErr = impl.RecursiveFetchSubmodules(rootDir)
		if cErr != nil {
			return errMsg, cErr
		}

		// cleanup

		if httpsAuth {
			_, errMsg, cErr = impl.UnsetCredentialHelper(rootDir)
			if cErr != nil {
				return errMsg, cErr
			}

			// delete file (~/.git-credentials) (which was created above)
			cErr = util.CleanupAfterFetchingHttpsSubmodules()
			if cErr != nil {
				return "", cErr
			}
		}
	}

	return "", nil

}
