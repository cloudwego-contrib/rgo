package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
)

func GenerateRGOCode(idlPath, repoPath string) error {
	fileType := filepath.Ext(idlPath)

	switch fileType {
	case ".thrift":
		err := generateThriftCode(idlPath, repoPath)
		if err != nil {
			return err
		}

		return generateClientCode(idlPath, repoPath)
	case ".proto":
		return nil
	default:
		return errors.New("unsupported idl file")
	}
}

func generateClientCode(idlPath, repoPath string) error {
	thriftFile, err := parseIDLFile(idlPath)
	if err != nil {
		return err
	}

	// Create a new file set
	fset := token.NewFileSet()

	namespace, f, err := buildThriftAstFile(thriftFile)
	if err != nil {
		return err
	}

	outputDir := filepath.Join(repoPath, "src/rgo-gen-go", namespace)

	// Generate the Go code
	outputFile, err := os.Create(filepath.Join(outputDir, "cli.go"))
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	if err := format.Node(outputFile, fset, f); err != nil {
		panic(err)
	}

	return nil
}

func buildThriftAstFile(thrift *parser.Thrift) (string, *ast.File, error) {
	// Create the AST for the Go file
	var namespace string

	for _, v := range thrift.Namespaces {
		if v.Language == "go" {
			namespace = v.Name
		}
	}

	if namespace == "" {
		return "", nil, errors.New("no go namespace found")
	}

	f := &ast.File{
		Name: ast.NewIdent(namespace),
	}

	f.Imports = []*ast.ImportSpec{
		{Path: &ast.BasicLit{Kind: token.STRING, Value: `"context"`}},
		{Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client"`}},
		{Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/cloudwego/kitex/client/callopt"`}},
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
										Type:  ast.NewIdent(s.Name),
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
								Type: &ast.ArrayType{
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
								Type: &ast.StarExpr{
									X: ast.NewIdent(s.Name + "Client"),
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
						X: ast.NewIdent(arg.Type.Name),
					},
				})
			}

			t = append(t, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent("opts")},
				Type: &ast.ArrayType{
					Elt: &ast.SelectorExpr{
						X:   ast.NewIdent("callopt"),
						Sel: ast.NewIdent("CallOptions"),
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
				Name: ast.NewIdent(function.Name),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: t,
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.StarExpr{
									X: ast.NewIdent(function.FunctionType.Name),
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
					Name: ast.NewIdent(function.Name),
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: t,
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.StarExpr{
										X: ast.NewIdent(function.FunctionType.Name),
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

	return namespace, f, nil
}
