package utils

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os/exec"
)

func CloneGitRepo(repoURL, branch, path, commit string) error {
	// clone repo
	cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", repoURL, path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone the repo: %v", err)
	}

	if commit == "" {
		return nil
	}

	// checkout commit
	cmd = exec.Command("git", "checkout", commit)
	cmd.Dir = path

	if err := cmd.Run(); err != nil {
		// fetch commit
		cmd = exec.Command("git", "fetch", "origin", commit)
		cmd.Dir = path

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to fetch the specified commit: %v", err)
		}

		cmd = exec.Command("git", "checkout", commit)
		cmd.Dir = path

		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout the specified commit: %v", err)
		}
	}

	return nil
}

func UpdateGitRepo(branch, path, commit string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	// Pull the latest changes from the remote branch
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Force:         true, // force to update in case of branch conflict
	})

	// If the repository is already up to date, continue without an error
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to pull the repo: %v", err)
	}

	// If commit is not empty, checkout to the specified commit
	if commit != "" {
		// Checkout to the specified commit
		err = w.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(commit),
			Force: true, // force to checkout to the specific commit
		})
		if err != nil {
			return fmt.Errorf("failed to checkout to commit %s: %v", commit, err)
		}
	}

	return nil
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
