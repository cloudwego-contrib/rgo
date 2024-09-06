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
	"path/filepath"
	"regexp"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/cloudwego-contrib/rgo/pkg/utils"

	"github.com/urfave/cli/v2"
)

const settingJson = `
{
  "go.toolsEnvVars": {
    "GOPACKAGESDRIVER":"${env:GOPATH}/bin/rgo_packages_driver"
  },
  "go.enableCodeLens": {
    "runtest": false
  },
  "gopls": {
    "formatting.gofumpt": true,
    "formatting.local": "rgo/",
    "ui.completion.usePlaceholders": false,
    "ui.semanticTokens": true,
    "ui.codelenses": {
      "gc_details": false,
      "regenerate_cgo": false,
      "generate": false,
      "test": false,
      "tidy": false,
      "upgrade_dependency": false,
      "vendor": false
    }
  },
  "go.useLanguageServer": true,
  "go.buildOnSave": "off",
  "go.lintOnSave": "off",
  "go.vetOnSave": "off"
}`

var re = regexp.MustCompile(`(?m)"go\.toolsEnvVars"\s*:\s*\{[^}]*"GOPACKAGESDRIVER"\s*:\s*"[^"]*"`)

func InitProject(c *cli.Context) error {
	workdir, err := os.Getwd()
	if err != nil {
		return err
	}
	// Create the directory structure of the project
	err = os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		return err
	}
	idetype := c.String(consts.TypeFlag)
	if idetype == "" {
		idetype = consts.VSCode
	}
	switch idetype {
	case consts.VSCode:
		// Create the .vscode directory
		return createVSCodeRGOConfig(workdir)
	}
	return os.WriteFile(filepath.Join(workdir, "rgo_config.yaml"), []byte("# "+filepath.Base(workdir)), os.ModePerm)
}

func createVSCodeRGOConfig(workdir string) error {
	err := os.MkdirAll(filepath.Join(workdir, consts.VSCodeDir), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create vscode directory: %v", err)
	}

	settingsFilePath := filepath.Join(workdir, consts.VSCodeDir, consts.SettingsJson)

	exist, err := utils.PathExist(settingsFilePath)
	if err != nil {
		return fmt.Errorf("failed to check vscode settings.json: %v", err)
	}

	if !exist {
		err = os.WriteFile(settingsFilePath, []byte(settingJson), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create vscode settings.json: %v", err)
		}
	} else {
		fileContent, err := os.ReadFile(settingsFilePath)
		if err != nil {
			return fmt.Errorf("Failed to read vscode settings.json: %v\n", err)
		}

		updatedContent := re.ReplaceAllStringFunc(string(fileContent), func(match string) string {
			return regexp.MustCompile(`"GOPACKAGESDRIVER"\s*:\s*"[^"]*"`).ReplaceAllString(match, `"GOPACKAGESDRIVER": "${env:GOPATH}/bin/rgo_packages_driver"`)
		})

		err = os.WriteFile(settingsFilePath, []byte(updatedContent), os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to write updated settings.json: %v\n", err)
		}
	}

	return nil
}
