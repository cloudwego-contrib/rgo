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

package tpl

import (
	"bytes"
	"text/template"
)

var anonymousDependenceTpl = `package main

import (
	{{ range $key, $value := . }}  
    _ "{{ $key }}"
	{{ end }}
)

func main() {}`

func RenderAnonymousDependenceTpl(m map[string]struct{}) (string, error) {
	buff := new(bytes.Buffer)
	tmpl, err := template.New("anon_dependence").Parse(anonymousDependenceTpl)
	if err != nil {
		return "", err
	}
	if err = tmpl.Execute(buff, m); err != nil {
		return "", err
	}
	return buff.String(), nil
}

type ClientGoTemplate struct {
	Imports         map[string]struct{}
	IdlSrvNameLower string
	IdlFileName     string
	Psm             string
}

var clientInitTpl = `package kitex_gen

import (
	{{ range $key, $value := .Imports }}  
    "{{ $key }}"
	{{ end }}
)

var {{.IdlFileName}}Client {{.IdlSrvNameLower}}.Client

func init() {
	c, err := {{.IdlSrvNameLower}}.NewClient("{{.Psm}}")
	if err != nil {
		panic(err)
	}
	{{.IdlFileName}}Client = c
}`

func (c *ClientGoTemplate) Render() (string, error) {
	return render("client_init_go", clientInitTpl, c)
}

type ClientFuncBody struct {
	ClientName      string
	FuncName        string
	FirstParamName  string
	SecondParamName string
	RealReqType     string
	NeedRespType    string
}

var clientFuncBodyTpl = `realResp, err := kitex_gen.{{.ClientName}}.{{.FuncName}}({{.FirstParamName}}, ({{.RealReqType}})(unsafe.Pointer({{.SecondParamName}})), callopt.WithHostPort("127.0.0.1:8888"))
return ({{.NeedRespType}})(unsafe.Pointer(realResp)), err`

func (cfb *ClientFuncBody) Render() (string, error) {
	return render("client_func_body_replace", clientFuncBodyTpl, cfb)
}

type SrvMainBody struct {
	PackageName string
	ServerImpl  string
}

var srvMainBodyTpl = `addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8888")
if err != nil {
	panic(err)
}

srv := {{.PackageName}}.NewServer(new({{.ServerImpl}}), server.WithServiceAddr(addr))
if err = srv.Run(); err != nil {
	panic(err)
}`

func (smb *SrvMainBody) Render() (string, error) {
	return render("service_main_body_replace", srvMainBodyTpl, smb)
}

type SrvHandlerBody struct {
	PackageName  string
	FuncName     string
	NeedReqType  string
	RealRespType string
}

var srvHandlerBodyTpl = `needReq := ({{.NeedReqType}})(unsafe.Pointer(req))
needResp, err := {{.PackageName}}.{{.FuncName}}(ctx, needReq)
return ({{.RealRespType}})(unsafe.Pointer(needResp)), err`

func (shb *SrvHandlerBody) Render() (string, error) {
	return render("service_handler_body_replace", srvHandlerBodyTpl, shb)
}

func render(name, text string, data any) (string, error) {
	buff := new(bytes.Buffer)
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", err
	}
	if err = tmpl.Execute(buff, data); err != nil {
		return "", err
	}
	return buff.String(), nil
}
