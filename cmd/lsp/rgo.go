package main

import (
	"github.com/cloudwego-contrib/rgo/pkg/generator"
	"github.com/cloudwego-contrib/rgo/pkg/global"
	config2 "github.com/cloudwego-contrib/rgo/pkg/global/config"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	idlConfigPath string

	c *config2.RGOConfig
	g *generator.RGOGenerator
)

func init() {
	idlConfigPath = consts.RGOConfigPath

	var err error

	rgoBasePath := os.Getenv(consts.RGOBasePath)
	if rgoBasePath == "" {
		rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), ".RGO", "cache")
	}

	currentPath, err := utils.GetCurrentPathWithUnderline()
	if err != nil {
		panic("get current path failed, err:" + err.Error())
	}

	global.InitLogger(rgoBasePath, currentPath)

	c, err = config2.ReadConfig(idlConfigPath)
	if err != nil {
		global.Logger.Warn("read rgo_config failed", zap.Error(err))
	}

	g = generator.NewRGOGenerator(c, rgoBasePath, currentPath)
}

func RGORun() {
	go WatchConfig(g)

	g.Run()
}

func WatchConfig(g *generator.RGOGenerator) {
	viper.WatchConfig()

	// hook function for config file change
	config2.ConfigChangeHandler = func(e fsnotify.Event) {
		viper.Reset()
		c, err := config2.ReadConfig(idlConfigPath)
		if err != nil {
			global.Logger.Error("read rgo_config failed", zap.Error(err))
		}

		global.Logger.Info("Config file changed:", zap.String("file_name", e.Name), zap.Any("config", c))

		g := generator.NewRGOGenerator(c, g.RGOBasePath, g.CurWorkPath)

		g.Run()
	}

	viper.OnConfigChange(config2.ConfigChangeHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
}