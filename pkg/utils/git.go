/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneGitRepo automatically selects HTTP or SSH based on the repoURL and clones the repository
func CloneGitRepo(path, commit string, opts *git.CloneOptions) error {
	_, err := git.PlainClone(path, false, opts)
	if err != nil {
		return fmt.Errorf("failed to clone the repo: %v", err)
	}

	// If a specific commit is provided, checkout that commit
	if commit != "" {
		repo, err := git.PlainOpen(path)
		if err != nil {
			return fmt.Errorf("failed to open the repo: %v", err)
		}

		wt, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get the worktree: %v", err)
		}

		err = wt.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(commit),
		})
		if err != nil {
			return fmt.Errorf("failed to checkout the specified commit: %v", err)
		}
	}

	return nil
}

// UpdateGitRepo pulls the latest changes from the remote branch and checks out to a specific commit if provided
func UpdateGitRepo(commit string, wt *git.Worktree, opts *git.PullOptions) error {
	err := wt.Pull(opts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull the latest changes: %v", err)
	}

	// If a specific commit is provided, checkout that commit
	if commit != "" {
		err = wt.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(commit),
			Force: true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout to commit %s: %v", commit, err)
		}
	}
	return nil
}

// GetLatestCommitID returns the latest commit ID of the given repository
func GetLatestCommitID(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	// Open the repository
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open the repo: %v", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get the HEAD reference: %v", err)
	}

	return ref.Hash().String(), nil
}
