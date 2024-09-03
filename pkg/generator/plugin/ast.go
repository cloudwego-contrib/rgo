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
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/cloudwego/thriftgo/parser"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func BuildRGOThriftAstFile(formatServiceName string, thrift *parser.Thrift) (*ast.File, error) {
	var namespace string

	for _, v := range thrift.Namespaces {
		if v.Language == "go" {
			namespace = strings.ToLower(v.GetName())
		}
	}

	if namespace == "" {
		return nil, errors.New("no go namespace found")
	}

	f := &ast.File{
		Name: ast.NewIdent(formatServiceName),
	}

	f.Decls = append([]ast.Decl{
		&ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"context"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client/callopt"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, filepath.Join(consts.RGOModuleName, formatServiceName, "kitex_gen", namespace))},
				},
			},
		},
	}, f.Decls...)

	// Create the client struct
	for _, s := range thrift.Services {
		f.Decls = append(f.Decls,
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: ast.NewIdent(s.Name + "Client"),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{ast.NewIdent(s.Name)},
										Type: &ast.SelectorExpr{
											X:   ast.NewIdent(namespace),
											Sel: ast.NewIdent(s.Name),
										},
									},
								},
							},
						},
					},
				},
			}, &ast.FuncDecl{
				Name: ast.NewIdent(fmt.Sprintf("New%sClient", s.Name)),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("serviceName")},
								Type:  ast.NewIdent("string"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("opts")},
								Type: &ast.Ellipsis{
									Elt: &ast.SelectorExpr{
										X:   ast.NewIdent("client"),
										Sel: ast.NewIdent("Option"),
									},
								},
							},
						},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: ast.NewIdent(s.Name + "Client"),
							},
							{
								Type: ast.NewIdent("error"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CompositeLit{
									Type: ast.NewIdent(fmt.Sprintf("%sClient", s.Name)),
								},
								ast.NewIdent("nil"),
							},
						},
					},
				},
			},
		)
	}

	// Create the client methods
	for _, s := range thrift.Services {
		for _, function := range s.Functions {
			var t []*ast.Field
			t = append(t, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("ctx")},
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("context"),
					Sel: ast.NewIdent("Context"),
				},
			})

			for _, arg := range function.Arguments {
				t = append(t, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(arg.Name)},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent(namespace),
							Sel: ast.NewIdent(arg.Type.Name),
						},
					},
				})
			}

			t = append(t, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("opts")},
				Type: &ast.Ellipsis{
					Elt: &ast.SelectorExpr{
						X:   ast.NewIdent("callopt"),
						Sel: ast.NewIdent("Option"),
					},
				},
			})

			f.Decls = append(f.Decls, &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("c")},
							Type: &ast.StarExpr{
								X: ast.NewIdent(s.Name + "Client"),
							},
						},
					},
				},
				Name: ast.NewIdent(cases.Title(language.Und).String(function.Name)),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: t,
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.StarExpr{
									X: &ast.SelectorExpr{
										X:   ast.NewIdent(namespace),
										Sel: ast.NewIdent(function.FunctionType.Name),
									},
								},
							},
							{
								Type: ast.NewIdent("error"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("nil"),
								ast.NewIdent("nil"),
							},
						},
					},
				},
			},
				&ast.FuncDecl{
					Name: ast.NewIdent(cases.Title(language.Und).String(function.Name)),
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: t,
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent(namespace),
											Sel: ast.NewIdent(function.FunctionType.Name),
										},
									},
								},
								{
									Type: ast.NewIdent("error"),
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									ast.NewIdent("nil"),
									ast.NewIdent("nil"),
								},
							},
						},
					},
				})
		}
	}

	return f, nil
}

func BuildRGOGenThriftAstFile(serviceName, formatServiceName string, thrift *parser.Thrift) (*ast.File, error) {
	var namespace string

	for _, v := range thrift.Namespaces {
		if v.Language == "go" {
			namespace = strings.ToLower(v.GetName())
		}
	}

	if namespace == "" {
		return nil, errors.New("no go namespace found")
	}

	name := strings.ToLower(thrift.Services[0].Name)

	f := &ast.File{
		Name: ast.NewIdent(formatServiceName),
	}

	f.Decls = append([]ast.Decl{
		&ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"context"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client/callopt"`},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, filepath.Join(consts.RGOModuleName, formatServiceName, "kitex_gen", namespace))},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, filepath.Join(consts.RGOModuleName, formatServiceName, "kitex_gen", namespace, name))},
				},
			},
		},
	}, f.Decls...)

	// Create the client struct
	for _, s := range thrift.Services {
		f.Decls = append(f.Decls,
			&ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent("defaultClient")},
						Type:  &ast.StarExpr{X: ast.NewIdent(fmt.Sprintf("%sClient", s.Name))},
					},
				},
			},
			&ast.FuncDecl{
				Name: ast.NewIdent("init"),
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{ast.NewIdent("defaultClient")},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.UnaryExpr{
									Op: token.AND,
									X:  &ast.CompositeLit{Type: ast.NewIdent(fmt.Sprintf("%sClient", s.Name))},
								},
							},
						},
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.SelectorExpr{
									X:   ast.NewIdent("defaultClient"),
									Sel: ast.NewIdent("Client"),
								},
								ast.NewIdent("_"),
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: ast.NewIdent(fmt.Sprintf("New%sClient", s.Name)),
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: fmt.Sprintf(`"%s"`, serviceName),
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: ast.NewIdent(fmt.Sprintf("%sClient", s.Name)),
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: []*ast.Field{
									{
										Type: &ast.SelectorExpr{
											X:   ast.NewIdent(name),
											Sel: ast.NewIdent("Client"),
										},
									},
								},
							},
						},
					},
				},
			}, &ast.FuncDecl{
				Name: ast.NewIdent(fmt.Sprintf("New%sClient", s.Name)),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("serviceName")},
								Type:  ast.NewIdent("string"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("opts")},
								Type: &ast.Ellipsis{
									Elt: &ast.SelectorExpr{
										X:   ast.NewIdent("client"),
										Sel: ast.NewIdent("Option"),
									},
								},
							},
						},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.SelectorExpr{
									X:   ast.NewIdent(name),
									Sel: ast.NewIdent("Client"),
								},
							},
							{
								Type: ast.NewIdent("error"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								ast.NewIdent("serviceClient"),
								ast.NewIdent("err"),
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: ast.NewIdent(fmt.Sprintf("%s.NewClient", name)),
									Args: []ast.Expr{
										ast.NewIdent("serviceName"),
										ast.NewIdent("opts..."),
									},
								},
							},
						},
						&ast.IfStmt{
							Cond: &ast.BinaryExpr{
								X:  ast.NewIdent("err"),
								Op: token.NEQ,
								Y:  ast.NewIdent("nil"),
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											ast.NewIdent("nil"),
											ast.NewIdent("err"),
										},
									},
								},
							},
						},
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("serviceClient"),
								ast.NewIdent("nil"),
							},
						},
					},
				},
			},
		)
	}

	// Create the client methods
	for _, s := range thrift.Services {
		for _, function := range s.Functions {
			var t []*ast.Field
			t = append(t, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("ctx")},
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("context"),
					Sel: ast.NewIdent("Context"),
				},
			})

			for _, arg := range function.Arguments {
				t = append(t, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(arg.Name)},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent(namespace),
							Sel: ast.NewIdent(arg.Type.Name),
						},
					},
				})
			}

			t = append(t, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("opts")},
				Type: &ast.Ellipsis{
					Elt: &ast.SelectorExpr{
						X:   ast.NewIdent("callopt"),
						Sel: ast.NewIdent("Option"),
					},
				},
			})

			f.Decls = append(f.Decls, &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("c")},
							Type: &ast.StarExpr{
								X: ast.NewIdent(fmt.Sprintf("%sClient", s.Name)),
							},
						},
					},
				},
				Name: ast.NewIdent(cases.Title(language.Und).String(function.Name)),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: t,
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.StarExpr{
									X: &ast.SelectorExpr{
										X:   ast.NewIdent(namespace),
										Sel: ast.NewIdent(function.FunctionType.Name),
									},
								},
							},
							{
								Type: ast.NewIdent("error"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								ast.NewIdent("res"),
								ast.NewIdent("err"),
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("c"),
										Sel: ast.NewIdent("Client." + cases.Title(language.Und).String(function.Name)),
									},
									Args: []ast.Expr{
										ast.NewIdent("ctx"),
										ast.NewIdent("req"),
										ast.NewIdent("opts..."),
									},
								},
							},
						},
						&ast.IfStmt{
							Cond: &ast.BinaryExpr{
								X:  ast.NewIdent("err"),
								Op: token.NEQ,
								Y:  ast.NewIdent("nil"),
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											ast.NewIdent("nil"),
											ast.NewIdent("err"),
										},
									},
								},
							},
						},
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("res"),
								ast.NewIdent("nil"),
							},
						},
					},
				},
			},
				&ast.FuncDecl{
					Name: ast.NewIdent(cases.Title(language.Und).String(function.Name)),
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: t,
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent(namespace),
											Sel: ast.NewIdent(function.FunctionType.Name),
										},
									},
								},
								{
									Type: ast.NewIdent("error"),
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									ast.NewIdent("res"),
									ast.NewIdent("err"),
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("defaultClient"),
											Sel: ast.NewIdent(cases.Title(language.Und).String(function.Name)),
										},
										Args: []ast.Expr{
											ast.NewIdent("ctx"),
											ast.NewIdent("req"),
											ast.NewIdent("opts..."),
										},
									},
								},
							},
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  ast.NewIdent("err"),
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{
												ast.NewIdent("nil"),
												ast.NewIdent("err"),
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									ast.NewIdent("res"),
									ast.NewIdent("nil"),
								},
							},
						},
					},
				})
		}
	}

	return f, nil
}
