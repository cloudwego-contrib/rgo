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

package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego-contrib/rgo/pkg/config"
	"github.com/cloudwego/thriftgo/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/thriftgo/plugin"
)

func strToPointer(str string) *string {
	return &str
}

func GetRGOPlugin(pluginType, pwd, projectModule, serviceName, formatServiceName string) (*RGOPlugin, error) {
	rgoPlugin := &RGOPlugin{
		Type:              pluginType,
		Pwd:               pwd,
		ProjectModule:     projectModule,
		ServiceName:       serviceName,
		FormatServiceName: formatServiceName,
	}

	return rgoPlugin, nil
}

type RGOPlugin struct {
	Type              string
	ProjectModule     string
	ServiceName       string
	FormatServiceName string
	Pwd               string
}

func (r *RGOPlugin) GetName() string {
	return r.ProjectModule
}

func (r *RGOPlugin) GetPluginParameters() []string {
	return nil
}

func (r *RGOPlugin) Invoke(req *plugin.Request) (res *plugin.Response) {
	switch r.Type {
	case consts.EditPeriod:
		return r.generateEditClientTemplateData(req)
	case consts.BuildPeriod:
		return r.generateBuildClientTemplateData(req)
	}
	return nil
}

func (r *RGOPlugin) generateEditClientTemplateData(req *plugin.Request) (res *plugin.Response) {
	formatServiceName := r.FormatServiceName
	serviceName := r.ServiceName

	thrift := req.AST

	for k := range thrift.Services {
		for i := range thrift.Services[k].Functions {
			thrift.Services[k].Functions[i].Name = cases.Title(language.Und).String(thrift.Services[k].Functions[i].Name)
		}
	}

	templateData, err := r.buildClientTemplateData(serviceName, formatServiceName, thrift)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to build client template data: %v", err)),
		}
	}

	// Render the client template using the extracted data
	renderedCode, err := RenderEditClientTemplate(templateData)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to render ast file: %v", err)),
		}
	}

	exist, err := utils.FileExistsInPath(r.Pwd, consts.GoMod)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(err.Error()),
		}
	}

	if !exist {
		err = os.MkdirAll(r.Pwd, os.ModePerm)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(fmt.Sprintf("failed to create directory: %v", err)),
			}
		}

		err = utils.InitGoMod(r.ProjectModule, r.Pwd)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(err.Error()),
			}
		}
	}

	outputFile, err := os.Create(filepath.Join(r.Pwd, "rgo_cli.go"))
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to create file: %v", err)),
		}
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(renderedCode)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to write render code to file: %v", err)),
		}
	}

	err = utils.RunGoModTidyInDir(r.Pwd)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to go mod tidy: %v", err)),
		}
	}

	return &plugin.Response{}
}

func (r *RGOPlugin) generateBuildClientTemplateData(req *plugin.Request) (res *plugin.Response) {
	formatServiceName := r.FormatServiceName
	serviceName := r.ServiceName

	thrift := req.AST

	for k := range thrift.Services {
		for i := range thrift.Services[k].Functions {
			thrift.Services[k].Functions[i].Name = cases.Title(language.Und).String(thrift.Services[k].Functions[i].Name)
		}
	}

	templateData, err := r.buildClientTemplateData(serviceName, formatServiceName, thrift)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to build client template data: %v", err)),
		}
	}

	// Render the client template using the extracted data
	renderedCode, err := RenderCompileClientTemplate(templateData)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to render ast file: %v", err)),
		}
	}

	exist, err := utils.FileExistsInPath(r.Pwd, consts.GoMod)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(err.Error()),
		}
	}

	if !exist {
		err = os.MkdirAll(r.Pwd, os.ModePerm)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(fmt.Sprintf("failed to create directory: %v", err)),
			}
		}

		err = utils.InitGoMod(r.ProjectModule, r.Pwd)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(err.Error()),
			}
		}
	}

	outputFile, err := os.Create(filepath.Join(r.Pwd, "rgo_cli.go"))
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to create file: %v", err)),
		}
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(renderedCode)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to write render code to file: %v", err)),
		}
	}

	err = utils.RunGoModTidyInDir(r.Pwd)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to go mod tidy: %v", err)),
		}
	}

	return &plugin.Response{}
}

func (r *RGOPlugin) buildClientTemplateData(serviceName, formatServiceName string, thriftFile *parser.Thrift) (*config.RGOClientTemplateData, error) {
	data := &config.RGOClientTemplateData{
		RGOModuleName:     r.ProjectModule,
		ServiceName:       serviceName,
		FormatServiceName: formatServiceName,
		Imports:           []string{"context", "github.com/cloudwego/kitex/client", "github.com/cloudwego/kitex/client/callopt"},
		Thrift:            thriftFile,
	}

	return data, nil
}
