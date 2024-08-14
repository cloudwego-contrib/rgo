package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

func CloneOrUpdateGitRepo(repoURL, branch, path string) error {
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

func GetLatestFileCommitTime(filePath string) (time.Time, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return time.Time{}, err
	}

	cmd := exec.Command("git", "log", "-1", "--format=%cd", "--date=iso", "--", absPath)
	cmd.Dir = filepath.Dir(filePath)

	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get the latest commit time: %v", err)
	}

	commitTimeStr := strings.TrimSpace(string(out))
	commitTime, err := time.Parse("2006-01-02 15:04:05 -0700", commitTimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit time: %v", err)
	}

	return commitTime, nil
}

func GetLatestCommitTime(filePaths []string) (time.Time, error) {
	var latestTime time.Time

	for _, file := range filePaths {
		commitTime, err := GetLatestFileCommitTime(file)
		if err != nil {
			return time.Time{}, err
		}

		if commitTime.After(latestTime) {
			latestTime = commitTime
		}
	}

	return latestTime, nil
}
