package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CloneGitRepo(repoURL, branch, path string) error {
	cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", "--depth", "1", repoURL, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone the repo: %v", err)
	}

	return nil
}

func UpdateGitRepo(repoURL, branch, path string) error {
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
	return nil
}

func GetLatestCommitID(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	//todo 判断是否最优
	cmd := exec.Command("git", "log", "-1", "--format=%H")
	cmd.Dir = absPath

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get the latest commit ID: %v", err)
	}

	return strings.TrimSpace(string(out)), nil
}
