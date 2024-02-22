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

package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/common/logs"
)

type Argument struct {
	// command-line arguments
	Verbose bool

	// go build args
	ChainName string
	ChainArgs []string

	Cwd       string
	TempDir   string
	GoMod     string
	GoModPath string
	ClientCfg string
	ServerCfg string
}

func (a *Argument) Init() error {
	flag.BoolVar(&a.Verbose, "verbose", false, "turn on verbose mode")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(),
			"rgo [-verbose]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if a.Verbose {
		logs.Log.Level = logs.LevelInfo
	} else {
		logs.Log.Level = logs.LevelWarn
	}

	goToolDir := os.Getenv("GOTOOLDIR")
	if goToolDir == "" {
		return errors.New("env key `GOTOOLDIR` not found")
	}
	if len(os.Args) < 2 {
		return errors.New("need to install rgo")
	}
	for i, arg := range os.Args[1:] {
		if goToolDir != "" && strings.HasPrefix(arg, goToolDir) {
			a.ChainName = arg
			if len(os.Args[1:]) > i+1 {
				a.ChainArgs = os.Args[i+2:]
			}
			break
		}
	}

	// get current dir
	cwd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("converting current directory to absolute path failed, err: %v", err)
	}
	a.Cwd = cwd

	// get tempDir
	a.TempDir = path.Join(cwd, "temp")

	return nil
}
