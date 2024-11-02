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

package generator

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

func (rg *RGOGenerator) CloneGitRepo(repoURL, branch, path, commit string) error {
	var auth transport.AuthMethod

	if strings.HasPrefix(repoURL, "git@") {
		// SSH authentication
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get the current user: %v", err)
		}

		sshKeyPath := filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa")

		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			return fmt.Errorf("failed to parse SSH key: %v", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	// Clone the repository using the appropriate authentication method
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Auth:          auth,
	}

	err := utils.CloneGitRepo(path, commit, cloneOptions)
	if err != nil {
		return err
	}
	return nil
}

// UpdateGitRepo pulls the latest changes from the remote branch and checks out to a specific commit if provided
func (rg *RGOGenerator) UpdateGitRepo(branch, path, commit string) error {
	// Open the repository from the specified path
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open the repo: %v", err)
	}

	// Get the remote URL to determine whether SSH or HTTP is being used
	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get the remote: %v", err)
	}
	remoteURL := remote.Config().URLs[0]

	// Set up SSH authentication if the remote URL uses SSH
	var auth transport.AuthMethod

	if strings.HasPrefix(remoteURL, "git@") {
		// SSH authentication
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get the current user: %v", err)
		}

		sshKeyPath := filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa")

		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			return fmt.Errorf("failed to parse SSH key: %v", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	// Get the worktree to perform git operations
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get the worktree: %v", err)
	}

	// Pull the latest changes with or without SSH authentication
	pullOptions := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Force:         true,
	}
	if auth != nil {
		pullOptions.Auth = auth
	}

	err = utils.UpdateGitRepo(commit, wt, pullOptions)
	if err != nil {
		return err
	}

	return nil
}
