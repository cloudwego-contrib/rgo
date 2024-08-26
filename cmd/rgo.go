package main

import (
	"github.com/cloudwego-contrib/rgo/generator"
	"github.com/cloudwego-contrib/rgo/global"
	"github.com/cloudwego-contrib/rgo/global/config"
	"github.com/cloudwego-contrib/rgo/global/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	idlConfigPath string

	c *config.RGOConfig
	g *generator.RGOGenerator
)

func init() {
	idlConfigPath = consts.RGOConfigPath

	var err error

	c, err = config.ReadConfig(idlConfigPath)
	if err != nil {
		panic("read rgo_config failed, err:" + err.Error())
	}

	rgoBasePath := os.Getenv(consts.RGOCachePath)
	if rgoBasePath == "" {
		//todo:目录命名
		rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), ".RGO", "cache")
	}

	currentPath, err := utils.GetCurrentPathWithUnderline()
	if err != nil {
		panic("get current path failed, err:" + err.Error())
	}

	global.InitLogger(rgoBasePath, currentPath)

	g = generator.NewRGOGenerator(c, rgoBasePath, currentPath)
}

func RGORun() {
	go WatchConfig(g)

	err := g.Run()
	if err != nil {
		global.Logger.Error("run rgo generator failed", zap.Error(err))
	}

}

func WatchConfig(g *generator.RGOGenerator) {
	viper.WatchConfig()

	// 定义回调函数
	config.ConfigChangeHandler = func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)

		viper.Reset()
		c, err := config.ReadConfig(idlConfigPath)
		if err != nil {
			panic("read rgo_config failed, err:" + err.Error())
		}

		log.Printf("Updated config: %v", c)

		g := generator.NewRGOGenerator(c, g.RGOBasePath, g.CurWorkPath)

		if err := g.Run(); err != nil {
			log.Printf("Failed to run generator with updated config: %v", err)
		}
	}

	viper.OnConfigChange(config.ConfigChangeHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
}
