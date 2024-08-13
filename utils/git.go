package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func IsGitURL(url string) bool {
	gitURLPattern := `^(https?|git|ssh|rsync|git@github\.com)\:\/\/|git@[^:]+:[^/]+\/[^/]+(\.git)?$`
	re := regexp.MustCompile(gitURLPattern)
	return re.MatchString(url)
}

func GetGitRepoName(repoURL string) (string, error) {
	repoURL = strings.TrimSuffix(repoURL, ".git")

	repoName := path.Base(repoURL)

	if repoName == "" {
		return "", fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	return repoName, nil
}

func CloneGitRepo(repoURL, branch, path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", repoURL, path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone the repo: %v", err)
		}
	} else {
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("path exists but is not a git repository")
		}

		cmd := exec.Command("git", "-C", path, "pull", "origin", branch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update the repo: %v", err)
		}
	}

	return nil
}
