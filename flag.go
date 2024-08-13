package main

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/generator"
	"github.com/cloudwego-contrib/rgo/utils"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var IDLConfigPath flagSlice

// 自定义类型 flagSlice 用于支持多个路径输入
type flagSlice []string

func (f *flagSlice) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *flagSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func Init() {
	IDLConfigPathStr := os.Getenv(consts.IDLConfigPath)
	if IDLConfigPathStr == "" {
		return
	}

	IDLConfigPath = strings.Split(IDLConfigPathStr, ",")

	rgoInfo := &config.RgoInfo{}
	rgoInfo.IDLConfig = make([]config.IDLConfig, 0)

	for _, path := range IDLConfigPath {
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read config file %s: %v", path, err)
		}

		var config *config.Config
		err = yaml.Unmarshal(content, &config)
		if err != nil {
			log.Fatalf("Failed to parse YAML file %s: %v", path, err)
		}

		rgoInfo.IDLConfig = append(rgoInfo.IDLConfig, config.RgoInfo.IDLConfig...)
	}

	rgoRepoPath := os.Getenv(consts.RGORepositoryPath)
	if rgoRepoPath == "" {
		rgoRepoPath = filepath.Join(utils.GetDefaultUserPath(), "RGO")
	}

	// 生成代码
	for _, idlConfig := range rgoInfo.IDLConfig {

		if utils.IsGitURL(idlConfig.Repository) {
			err := generator.GenerateRemoteRGOCode(&idlConfig, rgoRepoPath)
			if err != nil {
				log.Fatalf("Failed to clone repository %s: %v", idlConfig.Repository, err)
			}
		} else {
			path := filepath.Join(idlConfig.Repository, idlConfig.IDLPath)

			err := generator.GenerateRGOCode(path, rgoRepoPath)
			if err != nil {
				log.Fatalf("Failed to generate code for %s: %v", idlConfig.IDLPath, err)
			}
		}
	}
}
