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

package main

import (
	"os"
	"path/filepath"

	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
)

func Clean() error {
	InitConfig()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	exist, err := utils.FileExistsInPath(wd, "go.work")
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	removeModulePaths := make([]string, 0)

	for _, repo := range c.IDLRepos {
		buildPath := filepath.Join(rgoBasePath, consts.BuildPath, repo.RepoName, repo.Commit)

		for k := len(c.IDLs) - 1; k >= 0; k-- {
			if c.IDLs[k].RepoName == repo.RepoName {
				path := filepath.Join(buildPath, c.IDLs[k].FormatServiceName)

				removeModulePaths = append(removeModulePaths, path)

				c.IDLs = append(c.IDLs[:k], c.IDLs[k+1:]...)
			}
		}
	}

	return utils.RemoveModulesFromGoWork(filepath.Join(wd, "go.work"), removeModulePaths)
}
