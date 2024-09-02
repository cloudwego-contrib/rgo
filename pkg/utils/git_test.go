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
	"testing"
)

func TestCloneGitRepo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	err = CloneGitRepo("https://github.com/cloudwego/hertz.git", "develop", tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateGitRepo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	err = CloneGitRepo("https://github.com/cloudwego/hertz.git", "develop", tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
	err = UpdateGitRepo("develop", tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tempDir)
}

func TestGetLatestCommitID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	err = CloneGitRepo("https://github.com/cloudwego/hertz.git", "develop", tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
	s, err := GetLatestCommitID(tempDir)
	t.Log(s)
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tempDir)
}
