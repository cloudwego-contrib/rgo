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
	"context"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/rlog"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type RGOGenerator struct {
	RGOBasePath       string
	rgoConfig         *config.RGOConfig
	changedRepoCommit map[string]string
	wg                sync.WaitGroup
}

func NewRGOGenerator(rgoConfig *config.RGOConfig, rgoBasePath string) *RGOGenerator {
	return &RGOGenerator{
		RGOBasePath: rgoBasePath,
		rgoConfig:   rgoConfig,
	}
}

type Gen interface {
	GenRgoClientCode(serviceName, idlPath, rgoSrcPath string) error
	GenRgoBaseCode(idlPath, rgoSrcPath string) error
}

func (rg *RGOGenerator) Run() {
	rg.generateRepoCode()

	rg.generateSrcCode()
}

func (rg *RGOGenerator) generateRepoCode() {
	modifiedCommits := make([][2]string, 0)
	idlRepos := rg.rgoConfig.IDLRepos

	g, ctx := errgroup.WithContext(context.Background())

	for _, repo := range idlRepos {
		g.Go(func(repo config.IDLRepo) func() error {
			return func() error {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				rg.processRepo(repo, modifiedCommits)
				return nil
			}
		}(repo))
	}

	if err := g.Wait(); err != nil {
		rlog.Errorf("Failed to process all idl repos: %v", err)
	}

	for _, v := range modifiedCommits {
		rg.changedRepoCommit[v[0]] = v[1]
	}
}

func (rg *RGOGenerator) processRepo(repo config.IDLRepo, modifiedCommits [][2]string) {
	filePath := filepath.Join(rg.RGOBasePath, consts.IDLPath, repo.RepoName)

	exist, err := utils.PathExist(filePath)
	if err != nil {
		rlog.Errorf("Failed to check if path %s exists: %v", filePath, err)
		return
	}

	if repo.Commit == "" {
		commit, err := rg.cloneRemoteRepo(repo, filePath, repo.Commit)
		if err != nil {
			rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
			return
		}
		modifiedCommits = append(modifiedCommits, [2]string{repo.RepoName, commit})
		return
	}

	if !exist {
		commit, err := rg.cloneRemoteRepo(repo, filePath, repo.Commit)
		if err != nil {
			rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
			return
		}
		modifiedCommits = append(modifiedCommits, [2]string{repo.RepoName, commit})
	} else {
		id, err := utils.GetLatestCommitID(filePath)
		if err != nil {
			rlog.Errorf("Failed to get latest commit id for %s: %v", repo, err)
			return
		}

		if id != repo.Commit {
			commit, err := rg.updateRemoteRepo(repo, filePath, repo.Commit)
			if err != nil {
				rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
				return
			}
			modifiedCommits = append(modifiedCommits, [2]string{repo.RepoName, commit})
		}
	}
}

func (rg *RGOGenerator) generateSrcCode() {
	changedRepoCommit := rg.changedRepoCommit

	idls := rg.rgoConfig.IDLs
	for _, idl := range idls {
		if _, ok := changedRepoCommit[idl.RepoName]; !ok {
			continue
		}
		srcPath := filepath.Join(rg.RGOBasePath, consts.RepoPath, idl.FormatServiceName)

		idlPath := filepath.Join(rg.RGOBasePath, consts.IDLPath, idl.RepoName, idl.IDLPath)

		err := rg.GenerateRGOCode(idl.FormatServiceName, idlPath, srcPath)
		if err != nil {
			rlog.Errorf("Failed to generate rgo code for %s: %v", idl.ServiceName, err)
		}

	}
}

func (rg *RGOGenerator) cloneRemoteRepo(repo config.IDLRepo, path, commit string) (string, error) {
	var id string
	var err error

	err = utils.CloneGitRepo(repo.GitUrl, repo.Branch, path, commit)
	if err != nil {
		return "", err
	}

	id, err = utils.GetLatestCommitID(path)
	if err != nil {
		return "", err
	}

	return id, rg.updateRGORepoCommit(repo.RepoName, id)
}

func (rg *RGOGenerator) updateRemoteRepo(repo config.IDLRepo, path, commit string) (string, error) {
	var id string
	var err error

	err = utils.UpdateGitRepo(repo.Branch, path, commit)
	if err != nil {
		return "", err
	}

	id, err = utils.GetLatestCommitID(path)
	if err != nil {
		return "", err
	}

	return id, rg.updateRGORepoCommit(repo.RepoName, id)
}

func (rg *RGOGenerator) updateRGORepoCommit(repoName, newCommit string) error {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			rlog.Errorf("Failed to update commit for %s: %v\nStack Trace:\n%s", repoName, r, stackTrace)
		}
	}()

	repos := viper.Get("idl_repos").([]interface{})
	var res []config.IDLRepo

	for _, repo := range repos {
		idlRepo := repo.(map[string]interface{})

		if idlRepo["repo_name"] == repoName {
			idlRepo["commit"] = newCommit
		}

		res = append(res, config.IDLRepo{
			RepoName: idlRepo["repo_name"].(string),
			GitUrl:   idlRepo["git_url"].(string),
			Branch:   idlRepo["branch"].(string),
			Commit:   idlRepo["commit"].(string),
		})
	}

	return config.RewriteRGOConfig("idl_repos", res)
}
