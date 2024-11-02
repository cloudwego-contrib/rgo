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
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

func TestCloneGitRepo(t *testing.T) {
	var auth transport.AuthMethod

	repoURL := "https://github.com/cloudwego/hertz.git"

	if strings.HasPrefix(repoURL, "git@") {
		// SSH authentication
		currentUser, err := user.Current()
		if err != nil {
			t.Fatalf("failed to get the current user: %v", err)
		}

		sshKeyPath := filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa")

		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			t.Fatalf("failed to read SSH key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			t.Fatalf("failed to parse SSH key: %v", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	// Clone the repository using the appropriate authentication method
	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName("develop"),
		SingleBranch:  true,
		Auth:          auth,
	}

	err := CloneGitRepo("./tmp/hertz", "", cloneOptions)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateGitRepo(t *testing.T) {
	_, err := os.Stat("./tmp/hertz")
	if err != nil {
		if os.IsNotExist(err) {
			TestCloneGitRepo(t)
		}
		t.Fatal(err)
	}

	repo, err := git.PlainOpen("./tmp/hertz")
	if err != nil {
		t.Fatalf("failed to open the repo: %v", err)
	}

	// Get the remote URL to determine whether SSH or HTTP is being used
	remote, err := repo.Remote("origin")
	if err != nil {
		t.Fatalf("failed to get the remote: %v", err)
	}
	remoteURL := remote.Config().URLs[0]

	// Set up SSH authentication if the remote URL uses SSH
	var auth transport.AuthMethod

	if strings.HasPrefix(remoteURL, "git@") {
		// SSH authentication
		currentUser, err := user.Current()
		if err != nil {
			t.Fatal()
		}

		sshKeyPath := filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa")

		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			t.Fatalf("failed to read SSH key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			t.Fatalf("failed to parse SSH key: %v", err)
		}

		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	// Get the worktree to perform git operations
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get the worktree: %v", err)
	}

	// Pull the latest changes with or without SSH authentication
	pullOptions := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName("develop"),
		Force:         true,
		Auth:          auth,
	}

	err = UpdateGitRepo("", wt, pullOptions)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLatestCommitID(t *testing.T) {
	_, err := os.Stat("./tmp/hertz")
	if err != nil {
		if os.IsNotExist(err) {
			TestCloneGitRepo(t)
		}
		t.Fatal(err)
	}
	s, err := GetLatestCommitID("./tmp/hertz")
	t.Log(s)
	if err != nil {
		t.Fatal(err)
	}
}
