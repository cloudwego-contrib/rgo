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

package global

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"path/filepath"
	"time"
)

var Logger *zap.Logger

func InitLogger(rgoBasePath string) {
	currentTime := time.Now().Format("2006-01-02")
	logFileName := filepath.Join(rgoBasePath, consts.LogPath, fmt.Sprintf("%s.log", currentTime))

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	writeSyncer := zapcore.AddSync(lumberjackLogger)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapcore.InfoLevel,
	)

	Logger = zap.New(core, zap.AddCaller())
	defer Logger.Sync()

	Logger.Info("test log info")
}
