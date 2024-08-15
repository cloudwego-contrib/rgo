package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type RGOGenerator struct {
	ic          *IDLCache
	rgoRepoPath string
	rgoConfig   *config.RGOConfig
	mu          sync.Mutex
	idlMutex    map[string]*sync.Mutex
	wg          sync.WaitGroup
}

func NewRGOGenerator(rgoConfig *config.RGOConfig, rgoRepoPath string) *RGOGenerator {
	return &RGOGenerator{
		ic:          NewIDLCache(),
		rgoRepoPath: rgoRepoPath,
		rgoConfig:   rgoConfig,
		idlMutex:    make(map[string]*sync.Mutex),
	}
}

func (rg *RGOGenerator) Run() error {
	//err := rg.ic.LoadCache(filepath.Join(rg.rgoRepoPath, consts.IDLCacheFile))
	//if err != nil {
	//	return errors.New(fmt.Sprintf("Failed to load cache: %v\n", err))
	//}

	changedRepo := make(map[string]bool)

	idlRepos := rg.rgoConfig.IDLRepos

	//TODO: 设置并发数上限
	//TODO: 报错记录日志？
	for repo := range idlRepos {
		rg.wg.Add(1)

		go func(repo string) {
			defer rg.wg.Done()

			filePath := filepath.Join(rg.rgoRepoPath, consts.IDLRemotePath, repo)
			exist, err := utils.PathExist(filePath)
			if err != nil {
				return
			}

			if !exist || idlRepos[repo].Commit == "" {
				err = rg.cloneOrUpdateRemoteRepo(repo, idlRepos[repo].Repo, idlRepos[repo].Branch, filePath)
				if err != nil {
					log.Printf("Failed to clone repository %s: %v", repo, err)
				}
				changedRepo[repo] = true
			} else {
				id, err := utils.GetLatestCommitID(filePath)
				if err != nil {
					return
				}

				if id != idlRepos[repo].Commit {
					err = rg.cloneOrUpdateRemoteRepo(repo, idlRepos[repo].Repo, idlRepos[repo].Branch, filePath)
					if err != nil {
						log.Printf("Failed to clone repository %s: %v", repo, err)
					}
					changedRepo[repo] = true
				}
			}
		}(repo)
	}

	rg.wg.Wait()

	idls := rg.rgoConfig.IDLs
	for _, idl := range idls {
		if _, ok := changedRepo[idl.IDLRepo]; !ok {
			continue
		}

		rg.wg.Add(1)
		go func(idl config.IDL) {
			defer rg.wg.Done()
			path := filepath.Join(rg.rgoRepoPath, consts.IDLRemotePath, idl.IDLRepo, idl.IDLPath)

			err := GenerateRGOCode(idl.IDLRepo, path, rg.rgoRepoPath)
			if err != nil {
				log.Printf("Failed to generate code for %s: %v", idl, err)
				return
			}
		}(idl)
	}

	//var changedIDLConfigs []*config.IDLConfig
	//
	//for _, idlConfig := range rg.rgoInfo.IDLConfig {
	//	ic := idlConfig
	//	rg.wg.Add(1)
	//	rg.localRepoWg.Add(1)
	//
	//	go func() {
	//		defer rg.wg.Done()
	//		if utils.IsGitURL(ic.Repository) {
	//			rg.localRepoWg.Done()
	//			generatedFilePath, err := rg.updateRemoteRGOCode(&ic)
	//			if err != nil {
	//				log.Printf("Failed to clone repository %s: %v", ic.Repository, err)
	//			}
	//
	//			path := filepath.Join(generatedFilePath, ic.IDLPath)
	//
	//			files, err := getThriftIncludeFiles(path)
	//			if err != nil {
	//				return
	//			}
	//
	//			t, err := utils.GetLatestCommitTime(files)
	//			if err != nil {
	//				return
	//			}
	//
	//			rg.ic.Cache[path] = &cache{
	//				TimeStamp: time.Now(),
	//			}
	//
	//			if c, ok := rg.ic.Cache[path]; ok && c.TimeStamp.After(t) {
	//				return
	//			}
	//
	//			err = GenerateRGOCode(ic.Repository, path, rg.rgoRepoPath)
	//			if err != nil {
	//				log.Printf("Failed to generate code for %s: %v", ic, err)
	//				return
	//			}
	//
	//			return
	//		}
	//		defer rg.localRepoWg.Done()
	//
	//		path := filepath.Join(ic.Repository, ic.IDLPath)
	//
	//		if _, ok := rg.ic.Cache[path]; ok {
	//			changed, err := rg.ic.HashHasChanged(path)
	//			if err != nil {
	//				log.Printf("Failed to check if %s has changed: %v", path, err)
	//			}
	//			if changed {
	//				changedIDLConfigs = append(changedIDLConfigs, &ic)
	//			}
	//		} else {
	//			changedIDLConfigs = append(changedIDLConfigs, &ic)
	//			err := rg.ic.AddHash(path)
	//			if err != nil {
	//				log.Printf("Failed to add hash for %s: %v", path, err)
	//				return
	//			}
	//		}
	//	}()
	//
	//}
	//
	//rg.localRepoWg.Wait()
	//
	//for _, v := range changedIDLConfigs {
	//	path := filepath.Join(v.Repository, v.IDLPath)
	//	err = GenerateRGOCode(v.Repository, path, rg.rgoRepoPath)
	//	if err != nil {
	//		return errors.New(fmt.Sprintf("Failed to generate code for %s: %v", v, err))
	//	}
	//}
	//
	//rg.wg.Wait()
	//
	//err = rg.ic.SaveCache(filepath.Join(rg.rgoRepoPath, consts.IDLCacheFile))
	//if err != nil {
	//	return errors.New(fmt.Sprintf("Failed to save cache: %v", err))
	//}

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

func (rg *RGOGenerator) cloneOrUpdateRemoteRepo(repo, repoURL string, branch string, path string) error {
	var id string
	var err error

	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = utils.CloneGitRepo(repoURL, branch, path)
		if err != nil {
			return err
		}
	} else {
		err = utils.UpdateGitRepo(repoURL, branch, path)
		if err != nil {
			return err
		}
	}

	id, err = utils.GetLatestCommitID(path)
	if err != nil {
		return err
	}

	return rg.updateRGORepoCommit(repo, id)
}

func (rg *RGOGenerator) rewriteRGOConfig(key string, value any) error {
	//err := viper.ReadInConfig()
	//if err != nil {
	//	return errors.New(fmt.Sprintf("Failed to read config file: %v", err))
	//}

	viper.Set(key, value)

	err := viper.WriteConfig()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to write config file: %v", err))
	}
	return nil
}

func (rg *RGOGenerator) updateRGORepoCommit(repoName, newCommit string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to update commit for %s: %v", repoName, r)
		}
	}()

	repos := viper.Get("idl_repos").(map[string]interface{})

	idlRepo := repos[repoName].(map[string]interface{})

	repos[repoName] = config.IDLRepo{
		Repo:   idlRepo["repo"].(string),
		Branch: idlRepo["branch"].(string),
		Commit: newCommit,
	}

	return rg.rewriteRGOConfig("idl_repos", repos)
}
