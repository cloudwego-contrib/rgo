package generator

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/global"
	"github.com/cloudwego-contrib/rgo/global/config"
	"github.com/cloudwego-contrib/rgo/global/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RGOGenerator struct {
	RGOBasePath string
	CurWorkPath string
	rgoConfig   *config.RGOConfig
	mu          sync.Mutex
	idlMutex    map[string]*sync.Mutex
	wg          sync.WaitGroup
}

func NewRGOGenerator(rgoConfig *config.RGOConfig, rgoBasePath, curWorkPath string) *RGOGenerator {
	return &RGOGenerator{
		RGOBasePath: rgoBasePath,
		CurWorkPath: curWorkPath,
		rgoConfig:   rgoConfig,
		idlMutex:    make(map[string]*sync.Mutex),
	}
}

func (rg *RGOGenerator) Run() error {
	changedRepoCommit := make(map[string]string)

	idlRepos := rg.rgoConfig.IDLRepos

	//TODO: 设置并发数上限
	//TODO: 报错记录日志？
	// generate code for each idl repo
	for _, repo := range idlRepos {
		rg.wg.Add(1)

		go func(repo config.IDLRepo) {
			defer rg.wg.Done()

			curWorkPath := fmt.Sprintf("rgo_%s", rg.CurWorkPath)

			filePath := filepath.Join(rg.RGOBasePath, consts.IDLPath, curWorkPath, repo.RepoName)
			exist, err := utils.PathExist(filePath)
			if err != nil {
				return
			}

			if !exist || repo.Commit == "" {
				commit, err := rg.cloneOrUpdateRemoteRepo(repo.RepoName, repo.RepoGit, repo.Branch, filePath)
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
					commit, err := rg.cloneOrUpdateRemoteRepo(repo.RepoName, repo.RepoGit, repo.Branch, filePath)
					if err != nil {
						global.Logger.Error(fmt.Sprintf("Failed to clone or updaterepository %s: %v", repo, err))
						return
					}
					changedRepoCommit[repo.RepoName] = commit
				}
			}
		}(repo)
	}

	rg.wg.Wait()

	idls := rg.rgoConfig.IDLs
	for _, idl := range idls {
		if _, ok := changedRepoCommit[idl.IDLRepo]; !ok {
			continue
		}
		servicePath := filepath.Join(rg.RGOBasePath, consts.RepoPath, idl.ServiceName)

		commit := changedRepoCommit[idl.IDLRepo]

		commitPath := filepath.Join(servicePath, fmt.Sprintf("%s-%v", commit, time.Now().Format("2006-01-02")))

		srcPath := filepath.Join(commitPath, idl.ServiceName)

		curWorkPath := fmt.Sprintf("rgo_%s", rg.CurWorkPath)

		idlPath := filepath.Join(rg.RGOBasePath, consts.IDLPath, curWorkPath, idl.IDLRepo, idl.IDLPath)

		err := rg.generateRGOCode(rg.CurWorkPath, idl.ServiceName, idlPath, srcPath)
		if err != nil {
			log.Printf("Failed to generate code for %s: %v", idl, err)
			return err
		}

	}

	return nil
}

func (rg *RGOGenerator) getOrCreateMutex(repo string) *sync.Mutex {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	if _, exists := rg.idlMutex[repo]; !exists {
		rg.idlMutex[repo] = &sync.Mutex{}
	}
	return rg.idlMutex[repo]
}

func (rg *RGOGenerator) cloneOrUpdateRemoteRepo(repo, repoURL string, branch string, path string) (string, error) {
	var id string
	var err error

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = utils.CloneGitRepo(repoURL, branch, path)
		if err != nil {
			return "", err
		}
	} else {
		err = utils.UpdateGitRepo(repoURL, branch, path)
		if err != nil {
			return "", err
		}
	}

	id, err = utils.GetLatestCommitID(path)
	if err != nil {
		return "", err
	}

	return id, rg.updateRGORepoCommit(repo, id)
}

func (rg *RGOGenerator) updateRGORepoCommit(repoName, newCommit string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to update commit for %s: %v", repoName, r)
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
			RepoGit:  idlRepo["repo_git"].(string),
			Branch:   idlRepo["branch"].(string),
			Commit:   idlRepo["commit"].(string),
		})
	}

	return config.RewriteRGOConfig("idl_repos", res)
}
