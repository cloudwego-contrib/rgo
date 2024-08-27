package generator

import (
	"context"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global"
	config2 "github.com/cloudwego-contrib/rgo/pkg/global/config"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	utils2 "github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"path/filepath"
	"sync"
	"time"
)

type RGOGenerator struct {
	RGOBasePath       string
	CurWorkPath       string
	rgoConfig         *config2.RGOConfig
	changedRepoCommit map[string]string
	wg                sync.WaitGroup
}

func NewRGOGenerator(rgoConfig *config2.RGOConfig, rgoBasePath, curWorkPath string) *RGOGenerator {
	return &RGOGenerator{
		RGOBasePath:       rgoBasePath,
		CurWorkPath:       curWorkPath,
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
		rg.processRepo(repo.(config2.IDLRepo), changedRepoCommit)
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

func (rg *RGOGenerator) processRepo(repo config2.IDLRepo, changedRepoCommit map[string]string) {
	curWorkPath := fmt.Sprintf("rgo_%s", rg.CurWorkPath)
	filePath := filepath.Join(rg.RGOBasePath, consts.IDLPath, curWorkPath, repo.RepoName)

	exist, err := utils2.PathExist(filePath)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("Failed to check if path %s exists: %v", filePath, err))
		return
	}

	if !exist || repo.Commit == "" {
		commit, err := rg.cloneRemoteRepo(repo, filePath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to clone or update repository %s: %v", repo, err))
			return
		}
		changedRepoCommit[repo.RepoName] = commit
	} else {
		id, err := utils2.GetLatestCommitID(filePath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to get latest commit id for %s: %v", repo, err))
			return
		}

		if id != repo.Commit {
			commit, err := rg.updateRemoteRepo(repo, filePath)
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
		if _, ok := changedRepoCommit[idl.IDLRepo]; !ok {
			continue
		}
		servicePath := filepath.Join(rg.RGOBasePath, consts.RepoPath, idl.ServiceName)

		commit := changedRepoCommit[idl.IDLRepo]

		srcPath := filepath.Join(servicePath, fmt.Sprintf("%s-%v", commit, time.Now().Format("2006-01-02")))

		curWorkPath := fmt.Sprintf("rgo_%s", rg.CurWorkPath)

		idlPath := filepath.Join(rg.RGOBasePath, consts.IDLPath, curWorkPath, idl.IDLRepo, idl.IDLPath)

		err := rg.GenerateRGOCode(rg.CurWorkPath, idl.ServiceName, idlPath, srcPath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("Failed to generate rgo code for %s: %v", idl.ServiceName, err))
		}

	}
}

func (rg *RGOGenerator) cloneRemoteRepo(repo config2.IDLRepo, path string) (string, error) {
	var id string
	var err error

	err = utils2.CloneGitRepo(repo.RepoGit, repo.Branch, path)
	if err != nil {
		return "", err
	}

	id, err = utils2.GetLatestCommitID(path)
	if err != nil {
		return "", err
	}

	return id, rg.updateRGORepoCommit(repo.RepoName, id)
}

func (rg *RGOGenerator) updateRemoteRepo(repo config2.IDLRepo, path string) (string, error) {
	var id string
	var err error

	err = utils2.UpdateGitRepo(repo.RepoGit, repo.Branch, path)
	if err != nil {
		return "", err
	}

	id, err = utils2.GetLatestCommitID(path)
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
	var res []config2.IDLRepo

	for _, repo := range repos {
		idlRepo := repo.(map[string]interface{})

		if idlRepo["repo_name"] == repoName {
			idlRepo["commit"] = newCommit
		}

		res = append(res, config2.IDLRepo{
			RepoName: idlRepo["repo_name"].(string),
			RepoGit:  idlRepo["repo_git"].(string),
			Branch:   idlRepo["branch"].(string),
			Commit:   idlRepo["commit"].(string),
		})
	}

	return config2.RewriteRGOConfig("idl_repos", res)
}
