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

package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/common/consts"
)

type AstFuncInfo struct {
	FuncName string
	Params   []Param
	Returns  []Param
}

type Param struct {
	Name string
	Type string
}

func checkAndParseComment(s string, isServer bool) (result []string, err error) {
	index := -1
	if isServer {
		index = strings.Index(s, consts.RgoServer)
	} else {
		index = strings.Index(s, consts.RgoClient)
	}
	lastIndex := -1
	for i := index; i < len(s); i++ {
		if string(s[i]) == "\n" {
			lastIndex = i
		}
	}
	if lastIndex == -1 {
		lastIndex = len(s)
	}
	result = strings.Split(s[index:lastIndex], ":")
	if len(result) != 5 {
		return nil, fmt.Errorf("has grammar errors in %v", s)
	}
	return result, nil
}

func checkAndParseFunc(targetFunc *ast.FuncDecl) (*AstFuncInfo, error) {
	info := &AstFuncInfo{
		FuncName: targetFunc.Name.Name,
	}

	// check and parse params
	if len(targetFunc.Type.Params.List) != 2 {
		return nil, errors.New("must specify 2 params")
	}
	params := make([]Param, 2)
	for index, param := range targetFunc.Type.Params.List {
		if param.Names == nil {
			return nil, errors.New("must specify param name")
		}
		t := ""

		// context.Context
		if index == 0 {
			ctxType, ok := param.Type.(*ast.SelectorExpr)
			if !ok {
				return nil, errors.New("the first param must be context.Context")
			}
			ctxXType, ok := ctxType.X.(*ast.Ident)
			if !ok {
				return nil, errors.New("the first param must be context.Context")
			}
			t = ctxXType.Name + "." + ctxType.Sel.Name
			if t != "context.Context" {
				return nil, errors.New("the first param must be context.Context")
			}
		}

		// req
		if index == 1 {
			reqType, ok := param.Type.(*ast.StarExpr)
			if !ok {
				return nil, errors.New("the second param must be star type")
			}
			reqXType, ok := reqType.X.(*ast.SelectorExpr)
			if !ok {
				return nil, errors.New("the second param must be *{packageName}.{StructType}")
			}
			reqXXType, ok := reqXType.X.(*ast.Ident)
			if !ok {
				return nil, errors.New("the second param must be *{packageName}.{StructType}")
			}
			t = "*" + reqXXType.Name + "." + reqXType.Sel.Name
		}

		params[index] = Param{
			Name: param.Names[0].Name,
			Type: t,
		}
	}

	// check and parse returns
	if len(targetFunc.Type.Results.List) != 2 {
		return nil, errors.New("must specify 2 returns")
	}
	returns := make([]Param, 2)
	for index, param := range targetFunc.Type.Results.List {
		t := ""

		// resp
		if index == 0 {
			respType, ok := param.Type.(*ast.StarExpr)
			if !ok {
				return nil, errors.New("the first return param must be star type")
			}
			respXType, ok := respType.X.(*ast.SelectorExpr)
			if !ok {
				return nil, errors.New("the first return param must be *{packageName}.{StructType}")
			}
			respXXType, ok := respXType.X.(*ast.Ident)
			if !ok {
				return nil, errors.New("the first return param must be *{packageName}.{StructType}")
			}
			t = "*" + respXXType.Name + "." + respXType.Sel.Name
		}

		// req
		if index == 1 {
			errType, ok := param.Type.(*ast.Ident)
			if !ok {
				return nil, errors.New("the second return param must be error")
			}
			if errType.Name != "error" {
				return nil, errors.New("the second return param must be error")
			}
		}

		name := ""
		if param.Names != nil {
			name = param.Names[0].Name
		}
		returns[index] = Param{
			Name: name,
			Type: t,
		}
	}

	info.Params = params
	info.Returns = returns
	return info, nil
}
