package main

import (
	"github.com/cloudwego-contrib/rgo/pkg/generator"
	"github.com/cloudwego-contrib/rgo/pkg/global"
	"github.com/cloudwego-contrib/rgo/pkg/global/config"
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

	c *config.RGOConfig
	g *generator.RGOGenerator
)

func init() {
	idlConfigPath = consts.RGOConfigPath

	var err error

	currentPath, err := utils.GetCurrentPathWithUnderline()
	if err != nil {
		panic("get current path failed, err:" + err.Error())
	}

	rgoBasePath := filepath.Join(utils.GetDefaultUserPath(), consts.RGOBasePath, currentPath)

	global.InitLogger(rgoBasePath)

	c, err = config.ReadConfig(idlConfigPath)
	if err != nil {
		global.Logger.Warn("read rgo_config failed", zap.Error(err))
	}

	g = generator.NewRGOGenerator(c, rgoBasePath)
}

func RGORun() {
	go WatchConfig(g)

	g.Run()
}

func WatchConfig(g *generator.RGOGenerator) {
	viper.WatchConfig()

	// hook function for config file change
	config.ConfigChangeHandler = func(e fsnotify.Event) {
		viper.Reset()
		c, err := config.ReadConfig(idlConfigPath)
		if err != nil {
			global.Logger.Error("read rgo_config failed", zap.Error(err))
		}

		global.Logger.Info("Config file changed:", zap.String("file_name", e.Name), zap.Any("config", c))

		g := generator.NewRGOGenerator(c, g.RGOBasePath)

		g.Run()
	}

	viper.OnConfigChange(config.ConfigChangeHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
}
