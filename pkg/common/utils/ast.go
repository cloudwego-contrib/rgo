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

package utils

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
)

func AppendImports(fSet *token.FileSet, astFile *ast.File, imp []string) {
	addImports := make([]string, 0, len(imp))
	for _, i := range imp {
		flag := 0
		ast.Inspect(astFile, func(n ast.Node) bool {
			if importSpec, ok := n.(*ast.ImportSpec); ok && importSpec.Path.Value == i {
				flag = 1
				return false
			}
			return true
		})
		if flag == 0 {
			addImports = append(addImports, i)
		}
	}

	for _, i := range addImports {
		astutil.AddImport(fSet, astFile, i)
	}
}

func DeleteImport(fSet *token.FileSet, astFile *ast.File, imp string) {
	flag := false
	ast.Inspect(astFile, func(n ast.Node) bool {
		if importSpec, ok := n.(*ast.ImportSpec); ok && importSpec.Path.Value == imp {
			flag = true
			return false
		}
		return true
	})

	if flag == false {
		astutil.DeleteImport(fSet, astFile, imp)
	}
}

func ReplaceFuncBody(targetFunc *ast.FuncDecl, funcBody string) {
	targetFunc.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.BasicLit{
					Kind:  token.STRING,
					Value: funcBody,
				},
			},
		},
	}
}

func GetImportPath(pkgName string, astFile *ast.File) (string, error) {
	importValue := ""
	ast.Inspect(astFile, func(n ast.Node) bool {
		if importSpec, ok := n.(*ast.ImportSpec); ok {
			// remove the interference of colons
			if filepath.Base(importSpec.Path.Value[1:len(importSpec.Path.Value)-1]) == pkgName ||
				(importSpec.Name != nil && importSpec.Name.Name == pkgName) {
				importValue = importSpec.Path.Value[1 : len(importSpec.Path.Value)-1]
				return false
			}
		}
		return true
	})
	if importValue == "" {
		return "", fmt.Errorf("can not find %v's import path", pkgName)
	}
	return importValue, nil
}
