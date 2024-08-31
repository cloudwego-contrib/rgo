package generator

import (
	"context"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global"
	"github.com/cloudwego-contrib/rgo/pkg/global/config"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"path/filepath"
	"sync"
	"time"
)

type RGOGenerator struct {
	RGOBasePath       string
	rgoConfig         *config.RGOConfig
	changedRepoCommit map[string]string
	wg                sync.WaitGroup
}

func NewRGOGenerator(rgoConfig *config.RGOConfig, rgoBasePath string) *RGOGenerator {
	return &RGOGenerator{
		RGOBasePath:       rgoBasePath,
		rgoConfig:         rgoConfig,
		changedRepoCommit: make(map[string]string),
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
	changedRepoCommit := rg.changedRepoCommit

	idlRepos := rg.rgoConfig.IDLRepos

	// create errgroup
	g, ctx := errgroup.WithContext(context.Background())

	// create a gopool pool and limit the number of concurrent calls
	pool, _ := ants.NewPoolWithFunc(10, func(repo interface{}) {
		defer rg.wg.Done()
		rg.processRepo(repo.(config.IDLRepo), changedRepoCommit)
	})
	defer pool.Release()

	// traverse each idl repo and submit the task to gopool
	for _, repo := range idlRepos {
		repo := repo
		rg.wg.Add(1)

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return pool.Invoke(repo)
			}
		})
	}

	rg.wg.Wait()

	if err := g.Wait(); err != nil {
		global.Logger.Error(fmt.Sprintf("Failed to process all idl repos: %v", err))
	}
}

func (rg *RGOGenerator) processRepo(repo config.IDLRepo, changedRepoCommit map[string]string) {
	filePath := filepath.Join(rg.RGOBasePath, consts.IDLPath, repo.RepoName)

	exist, err := utils.PathExist(filePath)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("Failed to check if path %s exists: %v", filePath, err))
		return
	}

	if !exist {
		commit, err := rg.cloneRemoteRepo(repo, filePath, repo.Commit)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to clone or update repository %s: %v", repo, err))
			return
		}
		changedRepoCommit[repo.RepoName] = commit
	} else {
		id, err := utils.GetLatestCommitID(filePath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to get latest commit id for %s: %v", repo, err))
			return
		}

		if id != repo.Commit {
			commit, err := rg.updateRemoteRepo(repo, filePath, repo.Commit)
			if err != nil {
				global.Logger.Error(fmt.Sprintf("Failed to clone or update repository %s: %v", repo, err))
				return
			}
			changedRepoCommit[repo.RepoName] = commit
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
		servicePath := filepath.Join(rg.RGOBasePath, consts.RepoPath, idl.FormatServiceName)

		commit := changedRepoCommit[idl.RepoName]

		srcPath := filepath.Join(servicePath, fmt.Sprintf("%s-%v", commit, time.Now().Format("2006-01-02")))

		idlPath := filepath.Join(rg.RGOBasePath, consts.IDLPath, idl.RepoName, idl.IDLPath)

		err := rg.GenerateRGOCode(idl.FormatServiceName, idlPath, srcPath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to generate rgo code for %s: %v", idl.ServiceName, err))
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
			global.Logger.Error(fmt.Sprintf("Failed to update commit for %s: %v", repoName, r))
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
