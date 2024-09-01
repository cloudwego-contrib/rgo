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

package generator

import (
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"os"
	"path/filepath"

	"github.com/cloudwego-contrib/rgo/pkg/generator/plugin"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
)

func (rg *RGOGenerator) GenerateRGOCode(formatServiceName, idlPath, rgoSrcPath string) error {
	exist, err := utils.FileExistsInPath(rgoSrcPath, "go.mod")
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(rgoSrcPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		err = utils.InitGoMod(filepath.Join(consts.RGOModuleName, formatServiceName), rgoSrcPath)
		if err != nil {
			return err
		}
	}

	fileType := filepath.Ext(idlPath)

	switch fileType {
	case ".thrift":
		err = rg.GenRgoBaseCode(formatServiceName, idlPath, rgoSrcPath)
		if err != nil {
			return err
		}

		return rg.generateRGOPackages(formatServiceName, rgoSrcPath)
	case ".proto":
		return nil
	default:
		return errors.New("unsupported idl file")
	}
}

func (rg *RGOGenerator) GenRgoClientCode(serviceName, idlPath, rgoSrcPath string) error {
	thriftFile, err := parseIDLFile(idlPath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()

	f, err := plugin.BuildRGOThriftAstFile(serviceName, thriftFile)
	if err != nil {
		return err
	}

	outputDir := rgoSrcPath

	outputFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_cli.go", utils.GetFileNameWithoutExt(thriftFile.Filename))))
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if err = format.Node(outputFile, fset, f); err != nil {
		return err
	}

	return nil
}
