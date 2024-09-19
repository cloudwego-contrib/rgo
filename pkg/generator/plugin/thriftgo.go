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

func GetRGOThriftgoPlugin(pwd, projectModule, serviceName, formatServiceName string, Args []string) (*RGOThriftgoPlugin, error) {
	rgoPlugin := &RGOThriftgoPlugin{}

	rgoPlugin.Pwd = pwd
	rgoPlugin.ProjectModule = projectModule
	rgoPlugin.ServiceName = serviceName
	rgoPlugin.FormatServiceName = formatServiceName
	rgoPlugin.Args = Args

	return rgoPlugin, nil
}

type RGOThriftgoPlugin struct {
	Args              []string
	ProjectModule     string
	ServiceName       string
	FormatServiceName string
	Pwd               string
}

func (r *RGOThriftgoPlugin) GetName() string {
	return r.ProjectModule
}

func (r *RGOThriftgoPlugin) GetPluginParameters() []string {
	return r.Args
}

func (r *RGOThriftgoPlugin) Invoke(req *plugin.Request) (res *plugin.Response) {
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

func (r *RGOThriftgoPlugin) buildClientTemplateData(serviceName, formatServiceName string, thriftFile *parser.Thrift) (*config.RGOClientTemplateData, error) {
	data := &config.RGOClientTemplateData{
		RGOModuleName:     r.ProjectModule,
		ServiceName:       serviceName,
		FormatServiceName: formatServiceName,
		Imports:           []string{"context", "github.com/cloudwego/kitex/client", "github.com/cloudwego/kitex/client/callopt"},
		Thrift:            thriftFile,
	}

	return data, nil
}
