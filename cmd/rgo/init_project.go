package main

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

const settingJson = `
{
  "go.toolsEnvVars": {
    "GOPACKAGESDRIVER":"${GOPATH}/driver"
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
	modname := c.String("mod")
	if modname != "" {
		// Create the go.modname file
		err = utils.InitGoMod(modname, workdir)
		if err != nil {
			return err
		}
	} else {
		return errors.New("mod is required")
	}
	idetype := c.String("type")
	if idetype == "" {
		idetype = "vscode"
	}
	switch idetype {
	case "vscode":
		// Create the .vscode directory
		err = os.MkdirAll(filepath.Join(workdir, ".vscode"), os.ModePerm)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create vscode directory: %v\n", err))
		}
		err := os.WriteFile(filepath.Join(workdir, ".vscode", "settings.json"), []byte(settingJson), os.ModePerm)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create vscode settings.json: %v\n", err))
		}
	}
	return os.WriteFile(filepath.Join(workdir, "rgo_config.yaml"), []byte("# "+filepath.Base(workdir)), os.ModePerm)
}
