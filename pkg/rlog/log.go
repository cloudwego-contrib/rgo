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

package rlog

import (
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

func InitLogger(logPath string) {
	currentTime := time.Now().Format("2006-01-02")
	logFileName := filepath.Join(logPath, fmt.Sprintf("%s.log", currentTime))

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

	logger = zap.New(core, zap.AddCaller())
	defer logger.Sync()

	logger.Info("test log info")
}

func Debug(s string, fields ...zap.Field) {
	logger.Info(s, fields...)
}

func Info(s string, fields ...zap.Field) {
	logger.Info(s, fields...)
}

func Warn(s string, fields ...zap.Field) {
	logger.Warn(s, fields...)
}

func Error(s string, fields ...zap.Field) {
	logger.Error(s, fields...)
}

func Fatal(s string, fields ...zap.Field) {
	logger.Fatal(s, fields...)
}

func Debugf(format string, args ...interface{}) {
	logger.Sugar().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Sugar().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Sugar().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Sugar().Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Sugar().Fatalf(format, args...)
}
