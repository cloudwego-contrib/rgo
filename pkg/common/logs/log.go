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

package logs

import (
	"log"
	"os"
)

type Level int

const (
	LevelDebug = 1 + iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelStrMap = map[Level]string{
	LevelDebug: "rgo: [Debug]",
	LevelInfo:  "rgo: [Info]",
	LevelWarn:  "rgo: [Warn]",
	LevelError: "rgo: [Error]",
}

type LogFactory struct {
	Level Level
}

var Log = &LogFactory{Level: LevelWarn}

func logg(level Level, v ...any) {
	if Log.Level > level {
		return
	}
	if level == LevelError {
		log.Println(append([]any{levelStrMap[level]}, v...)...)
		os.Exit(2)
		return
	}
	log.Println(append([]any{levelStrMap[level]}, v...)...)
}

func Debug(v ...any) {
	logg(LevelDebug, v...)
}

func Info(v ...any) {
	logg(LevelInfo, v...)
}

func Warn(v ...any) {
	logg(LevelWarn, v...)
}

func Error(v ...any) {
	logg(LevelError, v...)
}
