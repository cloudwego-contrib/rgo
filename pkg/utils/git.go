package utils

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func CloneGitRepo(repoURL, branch, path string) error {
	file, err := findSSHPrivateKeyFile()
	if err != nil {
		return err
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", file, "")
	if err != nil {
		return fmt.Errorf("failed to generate public keys: %v", err)
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth:          publicKeys,
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone the repo: %v", err)
	}

	return nil
}

func UpdateGitRepo(branch, path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Force:         true, // if branch conflict, force to update
	})

	// if not update, ignore git.NoErrAlreadyUpToDate
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to pull the repo: %v", err)
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

func GetLatestCommitID(filePath string) (string, error) {
	r, err := git.PlainOpen(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open the repo: %v", err)
	}

	ref, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %v", err)
	}

	commitObj, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %v", err)
	}

	return commitObj.Hash.String(), nil
}

func findSSHPrivateKeyFile() (string, error) {
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	if homeDir == "" {
		return "", errors.New("could not find home directory")
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return "", fmt.Errorf("could not read .ssh dir %s: %w", sshDir, err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			return filepath.Join(sshDir, strings.TrimSuffix(file.Name(), ".pub")), nil
		}
	}
	return "", err
}
