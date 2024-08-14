package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"log"
	"path/filepath"
	"sync"
	"time"
)

type RGOGenerator struct {
	ic          *IDLCache
	rgoRepoPath string
	rgoInfo     *config.RgoInfo
	mu          sync.Mutex
	idlMutex    map[string]*sync.Mutex
	wg          sync.WaitGroup
	localRepoWg sync.WaitGroup
}

func NewRGOGenerator(rgoInfo *config.RgoInfo, rgoRepoPath string) *RGOGenerator {
	return &RGOGenerator{
		ic:          NewIDLCache(),
		rgoRepoPath: rgoRepoPath,
		rgoInfo:     rgoInfo,
		idlMutex:    make(map[string]*sync.Mutex),
	}
}

func (rg *RGOGenerator) Generate() error {
	err := rg.ic.LoadCache(filepath.Join(rg.rgoRepoPath, consts.IDLCacheFile))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load cache: %v\n", err))
	}

	var changedIDLConfigs []*config.IDLConfig

	//TODO: 设置并发数上限
	//TODO: 报错记录日志？
	for _, idlConfig := range rg.rgoInfo.IDLConfig {
		ic := idlConfig
		rg.wg.Add(1)
		rg.localRepoWg.Add(1)

		go func() {
			defer rg.wg.Done()
			if utils.IsGitURL(ic.Repository) {
				rg.localRepoWg.Done()
				generatedFilePath, err := rg.updateRemoteRGOCode(&ic)
				if err != nil {
					log.Printf("Failed to clone repository %s: %v", ic.Repository, err)
				}

				path := filepath.Join(generatedFilePath, ic.IDLPath)

				files, err := getThriftIncludeFiles(path)
				if err != nil {
					return
				}

				t, err := utils.GetLatestCommitTime(files)
				if err != nil {
					return
				}

				rg.ic.Cache[path] = &cache{
					TimeStamp: time.Now(),
				}

				if c, ok := rg.ic.Cache[path]; ok && c.TimeStamp.After(t) {
					return
				}

				err = GenerateRGOCode(ic.Repository, path, rg.rgoRepoPath)
				if err != nil {
					log.Printf("Failed to generate code for %s: %v", ic, err)
					return
				}

				return
			}
			defer rg.localRepoWg.Done()

			path := filepath.Join(ic.Repository, ic.IDLPath)

			if _, ok := rg.ic.Cache[path]; ok {
				changed, err := rg.ic.HashHasChanged(path)
				if err != nil {
					log.Printf("Failed to check if %s has changed: %v", path, err)
				}
				if changed {
					changedIDLConfigs = append(changedIDLConfigs, &ic)
				}
			} else {
				changedIDLConfigs = append(changedIDLConfigs, &ic)
				err := rg.ic.AddHash(path)
				if err != nil {
					log.Printf("Failed to add hash for %s: %v", path, err)
					return
				}
			}
		}()

	}

	rg.localRepoWg.Wait()

	for _, v := range changedIDLConfigs {
		path := filepath.Join(v.Repository, v.IDLPath)
		err = GenerateRGOCode(v.Repository, path, rg.rgoRepoPath)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to generate code for %s: %v", v, err))
		}
	}

	rg.wg.Wait()

	err = rg.ic.SaveCache(filepath.Join(rg.rgoRepoPath, consts.IDLCacheFile))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to save cache: %v", err))
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
