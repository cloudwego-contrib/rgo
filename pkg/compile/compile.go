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

package compile

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"
	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
	"github.com/cloudwego-contrib/rgo/pkg/parser"
)

func Compile(arg *cmd.Argument) error {
	files := make([]string, 0, len(arg.ChainArgs))
	goMod, goModPath, ok := utils.SearchGoMod(arg.Cwd, true)
	if !ok {
		return errors.New("can not find go mod")
	}
	arg.GoMod = goMod
	arg.GoModPath = goModPath

	packageName := ""
	for i, argument := range arg.ChainArgs {
		if argument == "-p" && i+1 < len(arg.ChainArgs) {
			packageName = arg.ChainArgs[i+1]
		}
		if strings.HasPrefix(argument, "-") {
			continue
		}
		if strings.Contains(argument, "importcfg") {
			arg.ServerCfg = argument
			arg.ClientCfg = argument
		}
		if strings.HasSuffix(argument, ".go") {
			files = arg.ChainArgs[i:]
			break
		}
	}

	if (packageName != "main" && !strings.HasPrefix(packageName, goMod)) || len(files) == 0 {
		return nil
	}

	// parse server
	if err := parser.ParseServer(files, arg); err != nil {
		return fmt.Errorf("parse server failed, err: %v", err)
	}
	// parse client
	if err := parser.ParseClient(files, arg); err != nil {
		return fmt.Errorf("parse client failed, err: %v", err)
	}

	return nil
}
