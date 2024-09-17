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
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/TobiasYin/go-lsp/lsp"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/rlog"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type RGOGenerator struct {
	isGoPackagesDriver bool
	RGOBasePath        string
	rgoConfig          *config.RGOConfig
	changedRepoCommit  *sync.Map
	LspServer          *lsp.Server
}

func NewRGOGenerator(lspServer *lsp.Server, rgoConfig *config.RGOConfig, rgoBasePath string) *RGOGenerator {
	var isGoPackagesDriver bool

	switch rgoConfig.Mode {
	case consts.GoPackagesDriverMode:
		isGoPackagesDriver = true
	case consts.GoWorkMode:
		isGoPackagesDriver = false
	default:
		rlog.Warn("unsupported rgo mode, use GoPackagesDriverMode as default", zap.String("mode", rgoConfig.Mode))
		isGoPackagesDriver = true
	}

	return &RGOGenerator{
		isGoPackagesDriver: isGoPackagesDriver,
		RGOBasePath:        rgoBasePath,
		rgoConfig:          rgoConfig,
		changedRepoCommit:  &sync.Map{},
		LspServer:          lspServer,
	}
}

type Gen interface {
	GenRgoClientCode(serviceName, idlPath, rgoSrcPath string) error
	GenRgoBaseCode(idlPath, rgoSrcPath string) error
}

func (rg *RGOGenerator) Run() {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			rlog.Errorf("Failed to run rgo: %v\nStack Trace:\n%s", r, stackTrace)
		}
	}()

	defer func() {
		err := rg.sendNotification(consts.MethodRGORestartLSP, nil)
		if err != nil {
			rlog.Errorf("Failed to restart LSP: %v", err)
			return
		}
	}()

	rg.generateRepoCode()

	rg.generateSrcCode()
}

func (rg *RGOGenerator) sendNotification(method string, params json.RawMessage) error {
	err := rg.LspServer.SendNotification(method, params)
	if err != nil {
		return err
	}
	return nil
}

func (rg *RGOGenerator) generateRepoCode() {
	idlRepos := rg.rgoConfig.IDLRepos

	var eg errgroup.Group

	for _, repo := range idlRepos {
		eg.Go(func(repo config.IDLRepo) func() error {
			return func() error {
				return rg.processRepo(repo, rg.changedRepoCommit)
			}
		}(repo))
	}

	if err := eg.Wait(); err != nil {
		rlog.Errorf("Failed to process all idl repos: %v", err)
		return
	}
}

func (rg *RGOGenerator) processRepo(repo config.IDLRepo, changedRepoCommit *sync.Map) error {
	filePath := filepath.Join(rg.RGOBasePath, consts.IDLPath, repo.RepoName)

	exist, err := utils.PathExist(filePath)
	if err != nil {
		rlog.Errorf("Failed to check if path %s exists: %v", filePath, err)
		return err
	}

	if repo.Commit == "" {
		err = os.RemoveAll(filePath)
		if err != nil {
			rlog.Errorf("Failed to remove repository %s: %v", repo, err)
			return err
		}

		commit, err := rg.cloneRemoteRepo(repo, filePath, repo.Commit)
		if err != nil {
			rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
			return err
		}
		changedRepoCommit.Store(repo.RepoName, commit)
		return nil
	}

	if !exist {
		commit, err := rg.cloneRemoteRepo(repo, filePath, repo.Commit)
		if err != nil {
			rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
			return err
		}
		changedRepoCommit.Store(repo.RepoName, commit)
	} else {
		id, err := utils.GetLatestCommitID(filePath)
		if err != nil {
			rlog.Errorf("Failed to get latest commit id for %s: %v", repo, err)
			return nil
		}

		if id != repo.Commit {
			commit, err := rg.updateRemoteRepo(repo, filePath, repo.Commit)
			if err != nil {
				rlog.Errorf("Failed to clone or update repository %s: %v", repo, err)
				return err
			}
			changedRepoCommit.Store(repo.RepoName, commit)
		}
	}
	return nil
}

func (rg *RGOGenerator) generateSrcCode() {
	changedRepoCommit := rg.changedRepoCommit

	if !rg.isGoPackagesDriver {
		wd, err := os.Getwd()
		if err != nil {
			rlog.Errorf("Failed to get current working directory: %v", err)
			return
		}

		exist, err := utils.FileExistsInPath(wd, consts.GoWork)
		if err != nil {
			rlog.Errorf("Failed to check if go.work exists in path %s: %v", wd, err)
			return
		}

		if !exist {
			err = utils.InitGoWork()
			if err != nil {
				rlog.Errorf("Failed to init go.work: %v", err)
				return
			}
			err = utils.AddModuleToGoWork(".")
			if err != nil {
				rlog.Errorf("Failed to add module to go.work: %v", err)
				return
			}
		}
	}

	idls := rg.rgoConfig.IDLs
	for _, idl := range idls {
		if _, ok := changedRepoCommit.Load(idl.RepoName); !ok {
			continue
		}
		srcPath := filepath.Join(rg.RGOBasePath, consts.RepoPath, idl.FormatServiceName)

		idlPath := filepath.Join(rg.RGOBasePath, consts.IDLPath, idl.RepoName, idl.IDLPath)

		err := rg.GenerateRGOCode(idl.ServiceName, idl.FormatServiceName, idlPath, srcPath)
		if err != nil {
			rlog.Errorf("Failed to generate rgo code for %s: %v", idl.ServiceName, err)
			return
		}

		if !rg.isGoPackagesDriver {
			err = utils.AddModuleToGoWork(srcPath)
			if err != nil {
				rlog.Errorf("Failed to add module to go.work: %v", err)
				return
			}
		}
	}

	err := utils.RunGoWorkSync()
	if err != nil {
		rlog.Errorf("Failed to run go work sync: %v", err)
		return
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
