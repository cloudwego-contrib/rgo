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
	"os"
	"path/filepath"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/generator/plugin"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/thriftgo/parser"
)

func (rg *RGOGenerator) GenerateRGOCode(serviceName, formatServiceName, idlPath, rgoSrcPath string) error {
	exist, err := utils.FileExistsInPath(rgoSrcPath, consts.GoMod)
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
	case consts.ThriftPostfix:
		err = rg.GenRgoBaseCode(serviceName, formatServiceName, idlPath, rgoSrcPath)
		if err != nil {
			return err
		}

		return rg.generatePackagesMeta(formatServiceName, rgoSrcPath)
	default:
		return errors.New("unsupported idl file: " + fileType)
	}
}

func (rg *RGOGenerator) GenRgoClientCode(serviceName, formatServiceName, idlPath, rgoSrcPath string) error {
	// Parse the IDL file to extract the thrift structure
	thriftFile, err := parseIDLFile(idlPath)
	if err != nil {
		return err
	}

	// Extract necessary information from the thrift file to create the template data
	templateData, err := rg.buildClientTemplateData(serviceName, formatServiceName, thriftFile)
	if err != nil {
		return err
	}

	// Render the client template using the extracted data
	renderedCode, err := plugin.RenderEditClientTemplate(templateData)
	if err != nil {
		return err
	}

	// Define the output directory and file path
	outputDir := rgoSrcPath
	outputFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_cli.go", utils.GetFileNameWithoutExt(thriftFile.Filename)))

	// Create the output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Write the rendered code to the output file
	_, err = outputFile.WriteString(renderedCode)
	if err != nil {
		return err
	}

	return nil
}

func (rg *RGOGenerator) buildClientTemplateData(serviceName, formatServiceName string, thriftFile *parser.Thrift) (*config.RGOClientTemplateData, error) {
	data := &config.RGOClientTemplateData{
		RGOModuleName:     consts.RGOModuleName,
		ServiceName:       serviceName,
		FormatServiceName: formatServiceName,
		Imports:           []string{"context", "github.com/cloudwego/kitex/client", "github.com/cloudwego/kitex/client/callopt"},
		Thrift:            thriftFile,
	}

	return data, nil
}
