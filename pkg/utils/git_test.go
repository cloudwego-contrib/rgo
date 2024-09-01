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
	err := CloneGitRepo("https://github.com/cloudwego/hertz.git", "develop", "./tmp/hertz", "")
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
	err = UpdateGitRepo("develop", "./tmp/hertz", "")
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
