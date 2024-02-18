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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"
	"github.com/cloudwego-contrib/rgo/pkg/common/logs"
	"github.com/cloudwego-contrib/rgo/pkg/compile"
)

func main() {
	arg := &cmd.Argument{}
	if err := arg.Init(); err != nil {
		logs.Error(err)
	}

	if arg.ChainName == "" {
		logs.Error("currently not in a compilation chain environment and rgo cannot be used")
	}

	toolName := filepath.Base(arg.ChainName)
	switch strings.TrimSuffix(toolName, ".exe") {
	case "compile":
		if err := compile.Compile(arg); err != nil {
			logs.Error(err)
		}
	case "link":
		if err := compile.Link(arg); err != nil {
			logs.Error(err)
		}
	}

	// build
	c := exec.Command(arg.ChainName, arg.ChainArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	if err := c.Run(); err != nil {
		logs.Error(fmt.Sprintf("run toolchain err, chainName: %v, err: %v", arg.ChainName, err))
	}
}

// Run flag
func Run() {}
