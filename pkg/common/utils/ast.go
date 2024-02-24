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
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

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

func GetFunction(funcName string, astFile *ast.File) *ast.FuncDecl {
	var funcDec *ast.FuncDecl
	ast.Inspect(astFile, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == funcName {
			funcDec = funcDecl
			return false
		}
		return true
	})
	return funcDec
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

func GetAstFileByStructName(searchPath, name string) (*ast.File, *ast.StructType, error) {
	files, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, nil, err
	}

	var file *ast.File
	var st *ast.StructType
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") {
			p := filepath.Join(searchPath, f.Name())
			fSet := token.NewFileSet()
			astFile, err := parser.ParseFile(fSet, p, nil, parser.ParseComments)
			if err != nil {
				return nil, nil, err
			}
			st = getStruct(name, astFile)
			if st != nil {
				file = astFile
				break
			}
		}
	}

	if st == nil {
		return nil, nil, fmt.Errorf("can not find struct %v", name)
	}

	return file, st, nil
}

func getStruct(name string, astFile *ast.File) *ast.StructType {
	var ty *ast.StructType
	ast.Inspect(astFile, func(n ast.Node) bool {
		if spec, ok := n.(*ast.TypeSpec); ok && spec.Name.Name == name {
			if t, ok := spec.Type.(*ast.StructType); ok {
				ty = t
				return false
			}
			return true
		}
		return true
	})
	return ty
}

func GetAstFileByFuncName(searchPath, name string) (*ast.File, error) {
	files, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, err
	}

	var file *ast.File
	var fun *ast.FuncDecl
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") {
			p := filepath.Join(searchPath, f.Name())
			fSet := token.NewFileSet()
			astFile, err := parser.ParseFile(fSet, p, nil, parser.ParseComments)
			if err != nil {
				return nil, err
			}
			fun = GetFunction(name, astFile)
			if fun != nil {
				file = astFile
				break
			}
		}
	}

	if fun == nil {
		return nil, fmt.Errorf("can not find function %v", name)
	}

	return file, nil
}
