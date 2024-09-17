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
	"bytes"
	"strings"
	"text/template"

	"github.com/cloudwego-contrib/rgo/pkg/config"
)

const defaultRGOEditClientTemplate = `package {{.FormatServiceName}}

import (
	{{- range .Imports }}
	"{{.}}"
	{{- end }}
	"{{.RGOModuleName}}/{{.FormatServiceName}}/kitex_gen/{{(index .Namespaces 0).Name}}"
)

type {{(index .Services 0).Name}}Client struct {
	{{(index .Services 0).Name}} {{(index .Namespaces 0).Name}}.{{(index .Services 0).Name}}
}

func New{{(index .Services 0).Name}}Client(serviceName string, opts ...client.Option) ({{(index .Services 0).Name}}Client, error) {
	return {{(index .Services 0).Name}}Client{}, nil
}

{{range (index .Services 0).Functions}}
func (c *{{(index $.Services 0).Name}}Client) {{.Name}}(ctx context.Context, {{range .Arguments}}{{.Name}} *{{(index $.Namespaces 0).Name}}.{{.Type}}, {{end}}opts ...callopt.Option) (*{{(index $.Namespaces 0).Name}}.{{.FunctionType}}, error) {
	return nil, nil
}

func {{.Name}}(ctx context.Context, {{range .Arguments}}{{.Name}} *{{(index $.Namespaces 0).Name}}.{{.Type}}, {{end}}opts ...callopt.Option) (*{{(index $.Namespaces 0).Name}}.{{.FunctionType}}, error) {
	return nil, nil
}
{{end}}
`

const defaultRGOCompileClientTemplate = `package {{.FormatServiceName}}

import (
	{{- range .Imports }}
	"{{.}}"
	{{- end }}
	"{{.RGOModuleName}}/{{.FormatServiceName}}/kitex_gen/{{(index .Namespaces 0).Name}}"
	"{{.RGOModuleName}}/{{.FormatServiceName}}/kitex_gen/{{(index .Namespaces 0).Name}}/{{ToLower (index .Services 0).Name}}"
)

var defaultClient *{{(index .Services 0).Name}}Client

func init() {
	defaultClient = &{{(index .Services 0).Name}}Client{}
	defaultClient.Client, _ = New{{(index .Services 0).Name}}Client("{{.ServiceName}}")
}

type {{(index .Services 0).Name}}Client struct {
	{{ToLower (index .Services 0).Name}}.Client
}

func New{{(index .Services 0).Name}}Client(serviceName string, opts ...client.Option) ({{ToLower (index .Services 0).Name}}.Client, error) {
	serviceClient, err := {{ToLower (index .Services 0).Name}}.NewClient(serviceName, opts...)
	if err != nil {
		return nil, err
	}
	return serviceClient, nil
}

{{range (index .Services 0).Functions}}
func (c *{{(index $.Services 0).Name}}Client) {{.Name}}(ctx context.Context, {{range .Arguments}}{{.Name}} *{{(index $.Namespaces 0).Name}}.{{.Type}}, {{end}}opts ...callopt.Option) (*{{(index $.Namespaces 0).Name}}.{{.FunctionType}}, error) {
	res, err := c.Client.{{.Name}}(ctx, {{range .Arguments}}{{.Name}}, {{end}}opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func {{.Name}}(ctx context.Context, {{range .Arguments}}{{.Name}} *{{(index $.Namespaces 0).Name}}.{{.Type}}, {{end}}opts ...callopt.Option) (*{{(index $.Namespaces 0).Name}}.{{.FunctionType}}, error) {
	res, err := defaultClient.{{.Name}}(ctx, {{range .Arguments}}{{.Name}}, {{end}}opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
{{end}}
`

func RenderEditClientTemplate(data *config.RGOClientTemplateData) (string, error) {
	tmpl, err := template.New("editClientTemplate").Parse(defaultRGOEditClientTemplate)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return "", err
	}

	return rendered.String(), nil
}

func RenderCompileClientTemplate(data *config.RGOClientTemplateData) (string, error) {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("compileClientTemplate").Funcs(funcMap).Parse(defaultRGOCompileClientTemplate)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return "", err
	}

	return rendered.String(), nil
}
