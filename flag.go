package main

import (
	"github.com/cloudwego-contrib/rgo/config"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/generator"
	"github.com/cloudwego-contrib/rgo/utils"
	"log"
	"os"
	"path/filepath"
)

func Init() {
	IDLConfigPath := os.Getenv(consts.IDLConfigPath)
	if IDLConfigPath == "" {
		IDLConfigPath = consts.RGOConfigDefaultPath
	}

	c, err := config.ReadConfig(IDLConfigPath)
	if err != nil {
		panic("read rgo_config failed, err:" + err.Error())
	}

	rgoRepoPath := os.Getenv(consts.RGORepositoryPath)
	if rgoRepoPath == "" {
		rgoRepoPath = filepath.Join(utils.GetDefaultUserPath(), "RGO")
	}

	g := generator.NewRGOGenerator(c, rgoRepoPath)

	err = g.Run()
	if err != nil {
		log.Println(err)
	}

}
