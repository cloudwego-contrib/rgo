package main

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global"
	"github.com/cloudwego-contrib/rgo/pkg/global/config"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/kitex/tool/cmd/kitex/sdk"
	"github.com/cloudwego/thriftgo/plugin"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

var (
	idlConfigPath string

	currentPath string
	rgoBasePath string

	c *config.RGOConfig
)

func InitConfig() {
	var err error

	rgoBasePath = os.Getenv(consts.RGOBasePath)
	if rgoBasePath == "" {
		rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), ".RGO", "cache")
	}

	currentPath, err = utils.GetCurrentPathWithUnderline()
	if err != nil {
		panic("get current path failed, err:" + err.Error())
	}

	currentPath = fmt.Sprintf("rgo_%s", currentPath)

	c, err = config.ReadConfig(idlConfigPath)
	if err != nil {
		global.Logger.Warn("read rgo_config failed", zap.Error(err))
	}

}

func GenerateRGOCode() error {
	InitConfig()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	modulePaths, err := utils.FindGoModDirectories(wd)
	if err != nil {
		return err
	}

	exist, err := utils.FileExistsInPath(wd, "go.work")
	if err != nil {
		return err
	}

	if !exist {
		err = utils.InitGoWork(modulePaths...)
		if err != nil {
			return err
		}
	}

	for _, repo := range c.IDLRepos {
		buildPath := filepath.Join(rgoBasePath, consts.BuildPath, currentPath, repo.RepoName, repo.Commit)

		for k := len(c.IDLs) - 1; k >= 0; k-- {
			if c.IDLs[k].RepoName == repo.RepoName {
				idlPath := filepath.Join(rgoBasePath, consts.IDLPath, currentPath, repo.RepoName, c.IDLs[k].IDLPath)

				path := filepath.Join(buildPath, c.IDLs[k].ServiceName)

				rgoPlugin, err := GetRGOPlugin(path, c.IDLs[k].ServiceName, nil)
				if err != nil {
					return err
				}

				err = generateKitexGen(path, filepath.Join("rgo", c.IDLs[k].ServiceName), idlPath, rgoPlugin)
				if err != nil {
					global.Logger.Error("Failed to generate rgo code", zap.Error(err))
				}

				err = utils.AddModuleToGoWork(path)
				if err != nil {
					return err
				}

				c.IDLs = append(c.IDLs[:k], c.IDLs[k+1:]...)
			}
		}
	}

	return nil
}

func generateKitexGen(wd, module, idlPath string, plugins ...plugin.SDKPlugin) error {
	err := sdk.RunKitexTool(wd, plugins, "--module", module, idlPath)
	if err != nil {
		return err
	}

	return nil
}
