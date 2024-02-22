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

type astFuncInfo struct {
	funcName string
	params   []param
	returns  []param
}

type param struct {
	isSamePkg bool
	name      string
	pType     string
}

func checkAndParseComment(s string) (result []string, err error) {
	index := -1
	index = strings.Index(s, consts.RgoClient)
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
	if len(result) != 2 {
		return nil, fmt.Errorf("has grammar errors in %v", s)
	}
	return result, nil
}

func checkAndParseFunc(targetFunc *ast.FuncDecl, pkgName string) (*astFuncInfo, error) {
	info := &astFuncInfo{
		funcName: targetFunc.Name.Name,
	}

	// check and parse params
	if len(targetFunc.Type.Params.List) != 2 {
		return nil, errors.New("must specify 2 params")
	}
	ps := make([]param, 2)
	for index, p := range targetFunc.Type.Params.List {
		if p.Names == nil {
			return nil, errors.New("must specify param name")
		}
		t := ""
		isSamePkg := false

		// context.Context
		if index == 0 {
			ctxType, ok := p.Type.(*ast.SelectorExpr)
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
			reqType, ok := p.Type.(*ast.StarExpr)
			if !ok {
				return nil, errors.New("the second param must be star type")
			}
			if reqXType, ok := reqType.X.(*ast.Ident); ok {
				// call within package
				t = "*" + pkgName + "." + reqXType.Name
				isSamePkg = true
			} else if reqXType, ok := reqType.X.(*ast.SelectorExpr); ok {
				// call outside of package, maybe local or remote
				reqXXType, ok := reqXType.X.(*ast.Ident)
				if !ok {
					return nil, errors.New("the second param must be *{packageName}.{StructType}")
				}
				t = "*" + reqXXType.Name + "." + reqXType.Sel.Name
			} else {
				return nil, fmt.Errorf("unsupported struct type in function %v", info.funcName)
			}
		}

		ps[index] = param{
			name:      p.Names[0].Name,
			pType:     t,
			isSamePkg: isSamePkg,
		}
	}

	// check and parse returns
	if len(targetFunc.Type.Results.List) != 2 {
		return nil, errors.New("must specify 2 returns")
	}
	returns := make([]param, 2)
	for index, p := range targetFunc.Type.Results.List {
		t := ""
		isSamePkg := false

		// resp
		if index == 0 {
			respType, ok := p.Type.(*ast.StarExpr)
			if !ok {
				return nil, errors.New("the first return param must be star type")
			}
			if respXType, ok := respType.X.(*ast.Ident); ok {
				// call within package
				t = "*" + pkgName + "." + respXType.Name
				isSamePkg = true
			} else if respXType, ok := respType.X.(*ast.SelectorExpr); ok {
				// call outside of package, maybe local or remote
				respXXType, ok := respXType.X.(*ast.Ident)
				if !ok {
					return nil, errors.New("the first return param must be *{packageName}.{StructType}")
				}
				t = "*" + respXXType.Name + "." + respXType.Sel.Name
			} else {
				return nil, fmt.Errorf("unsupported struct type in function %v", info.funcName)
			}
		}

		// error
		if index == 1 {
			errType, ok := p.Type.(*ast.Ident)
			if !ok {
				return nil, errors.New("the second return param must be error")
			}
			if errType.Name != "error" {
				return nil, errors.New("the second return param must be error")
			}
		}

		name := ""
		if p.Names != nil {
			name = p.Names[0].Name
		}
		returns[index] = param{
			name:      name,
			pType:     t,
			isSamePkg: isSamePkg,
		}
	}

	info.params = ps
	info.returns = returns
	return info, nil
}
