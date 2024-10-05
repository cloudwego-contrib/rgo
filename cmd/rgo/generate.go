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
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/kitex/tool/cmd/kitex/sdk"
	"github.com/cloudwego/thriftgo/plugin"
)

var (
	idlConfigPath string
	currentPath   string
	rgoBasePath   string

	kitexCustomArgs cli.StringSlice

	c *config.RGOConfig

	isGoPackagesDriver bool
)

func InitConfig() error {
	var err error

	currentPath, err = utils.GetProjectHashPathWithUnderline()
	if err != nil {
		panic(err)
	}

	rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), consts.RGOBasePath, currentPath)

	c, err = config.ReadConfig(idlConfigPath)
	if err != nil {
		panic(err)
	}

	isGoPackagesDriver = c.Mode == consts.GoPackagesDriverMode

	switch c.Mode {
	case consts.GoPackagesDriverMode:
		isGoPackagesDriver = true
	case consts.GoWorkMode:
		isGoPackagesDriver = false
	default:
		isGoPackagesDriver = true
		fmt.Println("warning: unsupported rgo mode, use GoPackagesDriverMode as default")
	}

	return nil
}

var g errgroup.Group

func GenerateRGOCode() error {
	if err := InitConfig(); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	exist, err := utils.FileExistsInPath(wd, consts.GoWork)
	if err != nil {
		return err
	}

	if !exist {
		err = utils.InitGoWork()
		if err != nil {
			return err
		}

		err = utils.AddModuleToGoWork(".")
		if err != nil {
			return err
		}
	}

	for _, repo := range c.IDLRepos {
		buildPath := filepath.Join(rgoBasePath, consts.BuildPath, repo.RepoName, repo.Commit)

		for k := len(c.IDLs) - 1; k >= 0; k-- {
			if c.IDLs[k].RepoName == repo.RepoName {

				idl := c.IDLs[k]
				repo := repo
				k := k

				g.Go(func() error {
					idlPath := filepath.Join(rgoBasePath, consts.IDLPath, repo.RepoName, idl.IDLPath)

					path := filepath.Join(buildPath, idl.FormatServiceName)

					module := strings.ReplaceAll(c.ProjectModule, consts.RGOServiceName, idl.FormatServiceName)

					args := []string{
						"kitex",
						fmt.Sprintf("--%s", consts.PwdFlag), path,
						fmt.Sprintf("--%s", consts.ModuleFlag), module,
						fmt.Sprintf("--%s", consts.ServiceNameFlag), c.IDLs[k].ServiceName,
						fmt.Sprintf("--%s", consts.FormatServiceNameFlag), c.IDLs[k].FormatServiceName,
						fmt.Sprintf("--%s", consts.IDLPathFlag), idlPath,
					}

					for _, customArg := range kitexCustomArgs.Value() {
						args = append(args, fmt.Sprintf("--%s", consts.KitexArgsFlag), customArg)
					}

					cmd := exec.Command("rgo", args...)

					if err = cmd.Run(); err != nil {
						return fmt.Errorf("error generate rgo kitex_gen code: %v", err)
					}

					if isGoPackagesDriver {
						err = utils.AddModuleToGoWork(path)
						if err != nil {
							return err
						}
					} else {
						oldPath := filepath.Join(rgoBasePath, consts.RepoPath, idl.FormatServiceName)

						err = utils.ReplaceModulesInGoWork(oldPath, path)
						if err != nil {
							return err
						}
					}

					return nil
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
		return err
	} else {
		return utils.RunGoWorkSync()
	}
}

func generateKitexGen(wd, module, idlPath string, customArgs []string, plugins ...plugin.SDKPlugin) error {
	err := sdk.RunKitexTool(wd, plugins, append(customArgs, "--module", module, idlPath)...)
	if err != nil {
		return err
	}

	return nil
}
