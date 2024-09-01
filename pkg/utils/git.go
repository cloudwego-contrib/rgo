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
	"os/exec"
	"path/filepath"
	"strings"
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
	// Change directory to the repository path
	cmd := exec.Command("git", "-C", path, "pull", "origin", branch, "--force")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull the repo: %v", err)
	}

	// If commit is not empty, checkout to the specified commit
	if commit != "" {
		cmd = exec.Command("git", "-C", path, "checkout", commit, "--force")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout to commit %s: %v", commit, err)
		}
	}

	return nil
}

func GetLatestCommitID(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "log", "-1", "--format=%H")
	cmd.Dir = absPath

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get the latest commit ID: %v", err)
	}

	return strings.TrimSpace(string(out)), nil
}
