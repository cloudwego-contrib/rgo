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
	"context"
	"path/filepath"
	"runtime/debug"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/generator"
	"github.com/cloudwego-contrib/rgo/pkg/rlog"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type App struct {
	idlConfigPath string
	rgoBasePath   string
	config        *config.RGOConfig
	generator     *generator.RGOGenerator
	isRunning     chan struct{} // judge whether the app is running
}

func NewApp() (*App, error) {
	idlConfigPath := consts.RGOConfigPath

	currentPath, err := utils.GetProjectHashPathWithUnderline()
	if err != nil {
		return nil, err
	}

	rgoBasePath := filepath.Join(utils.GetDefaultUserPath(), consts.RGOBasePath, currentPath)

	rlog.InitLogger(rgoBasePath)

	c, err := config.ReadConfig(idlConfigPath)
	if err != nil {
		rlog.Warn("read rgo_config failed", zap.Error(err))
	}

	g := generator.NewRGOGenerator(c, rgoBasePath)

	return &App{
		idlConfigPath: idlConfigPath,
		rgoBasePath:   rgoBasePath,
		config:        c,
		generator:     g,
	}, nil
}

func (app *App) Run(ctx context.Context) {
	app.isRunning <- struct{}{}
	defer func() {
		<-app.isRunning
	}()

	go app.WatchConfig(ctx)

	app.generator.Run()
}

func (app *App) WatchConfig(ctx context.Context) {
	viper.WatchConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		select {
		case app.isRunning <- struct{}{}:
			defer func() {
				<-app.isRunning
			}()
		default:
			rlog.Warn("A config change is already being processed, skipping this event.")
			return
		}

		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				rlog.Error("Recovered from panic in ConfigChangeHandler", zap.Any("error", r), zap.String("stack_trace", stackTrace))
			}
		}()

		viper.Reset()
		c, err := config.ReadConfig(app.idlConfigPath)
		if err != nil {
			rlog.Error("read rgo_config failed", zap.Error(err))
			return
		}

		rlog.Info("Config file changed:", zap.String("file_name", e.Name), zap.Any("config", c))

		generator.NewRGOGenerator(c, app.rgoBasePath).Run()
	})

	<-ctx.Done()
}
