package main

import (
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"os"
	"path/filepath"
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
