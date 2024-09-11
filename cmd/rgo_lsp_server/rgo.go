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

var isRunning = make(chan struct{}, 1)

func initConfig() *generator.RGOGenerator {
	var err error

	currentPath, err := utils.GetProjectHashPathWithUnderline()
	if err != nil {
		panic(err)
	}

	rgoBasePath := filepath.Join(utils.GetDefaultUserPath(), consts.RGOBasePath, currentPath)

	rlog.InitLogger(rgoBasePath)

	c, err := config.ReadConfig(consts.RGOConfigPath)
	if err != nil {
		rlog.Warn("read rgo_config failed, file not found", zap.Error(err))
	}

	return generator.NewRGOGenerator(c, rgoBasePath)
}

func RGORun(ctx context.Context) {
	g := initConfig()

	isRunning <- struct{}{}
	defer func() {
		<-isRunning
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				rlog.Error("Recovered from panic in WatchConfig goroutine", zap.Any("error", r), zap.String("stack_trace", stackTrace))
			}
		}()

		WatchConfig(g, ctx)
	}()

	g.Run()
}

func WatchConfig(g *generator.RGOGenerator, ctx context.Context) {
	viper.WatchConfig()

	viper.OnConfigChange(func(e fsnotify.Event) {
		select {
		case isRunning <- struct{}{}:
			defer func() {
				<-isRunning
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
		c, err := config.ReadConfig(consts.RGOConfigPath)
		if err != nil {
			rlog.Error("read rgo_config failed, file not found", zap.Error(err))
			return
		}

		rlog.Info("Config file changed:", zap.String("file_name", e.Name), zap.Any("config", c))

		generator.NewRGOGenerator(c, g.RGOBasePath).Run()
	})

	viper.OnConfigChange(config.ConfigChangeHandler)

	<-ctx.Done()
}
