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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"
	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
)

func Link(arg *cmd.Argument) error {
	var cfg string
	buildmode := false
	for _, argument := range arg.ChainArgs {
		if argument == "-buildmode=exe" ||
			// windows
			argument == "-buildmode=pie" {
			buildmode = true
		}
		if strings.HasPrefix(argument, "-") {
			continue
		}
		if strings.Contains(argument, filepath.Join("b001", "importcfg.link")) {
			cfg = argument
		}
	}

	if !buildmode || cfg == "" {
		return nil
	}

	// merge server import config
	basePath := filepath.Join(arg.TempDir, "depen", "server")
	genTxtPath := filepath.Join(basePath, "depency.txt")
	isExist, err := utils.PathExist(genTxtPath)
	if err != nil {
		return err
	}
	if isExist {
		if err = utils.MergeImportCfg(genTxtPath, cfg, "importcfg.link", false); err != nil {
			return fmt.Errorf("merge link phase server imports dependence failed, err: %v", err)
		}
	}

	// merge client import config
	basePath = filepath.Join(arg.TempDir, "depen", "client")
	genTxtPath = filepath.Join(basePath, "depency.txt")
	isExist, err = utils.PathExist(genTxtPath)
	if err != nil {
		return err
	}
	if isExist {
		if err = utils.MergeImportCfg(genTxtPath, cfg, "importcfg.link", false); err != nil {
			return fmt.Errorf("merge link phase client imports dependence failed, err: %v", err)
		}
	}

	return nil
}
