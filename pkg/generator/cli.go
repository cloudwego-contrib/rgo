package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/thriftgo/parser"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (rg *RGOGenerator) GenerateRGOCode(curWorkPath, serviceName, idlPath, rgoSrcPath string) error {
	exist, err := utils.FileExistsInPath(rgoSrcPath, "go.mod")
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(rgoSrcPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		err = initGoMod(filepath.Join(consts.RGOModuleName, serviceName), rgoSrcPath)
		if err != nil {
			return err
		}
	}

	fileType := filepath.Ext(idlPath)

	switch fileType {
	case ".thrift":
		err := GenRgoBaseCode(idlPath, rgoSrcPath)
		if err != nil {
			return err
		}

		err = rg.GenRgoClientCode(serviceName, idlPath, rgoSrcPath)
		if err != nil {
			return err
		}

		return rg.generateRGOPackages(curWorkPath, serviceName, rgoSrcPath)
	case ".proto":
		return nil
	default:
		return errors.New("unsupported idl file")
	}
}

func (rg *RGOGenerator) GenRgoClientCode(serviceName, idlPath, rgoSrcPath string) error {
	thriftFile, err := parseIDLFile(idlPath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()

	f, err := buildThriftAstFile(serviceName, thriftFile)
	if err != nil {
		return err
	}

	outputDir := rgoSrcPath

	outputFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_cli.go", utils.GetFileNameWithoutExt(thriftFile.Filename))))
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if err = format.Node(outputFile, fset, f); err != nil {
		return err
	}

	return nil
}

func initGoMod(moduleName, path string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)

	cmd.Dir = path

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to initialize go.mod in path '%s': %w", path, err)
	}

	// Set Go version to 1.18
	cmd = exec.Command("go", "mod", "edit", "-go=1.18")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set Go version 1.18 in path '%s': %w", path, err)
	}

	return nil
}

func buildThriftAstFile(serviceName string, thrift *parser.Thrift) (*ast.File, error) {
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
		Name: ast.NewIdent(serviceName),
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
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, filepath.Join(consts.RGOModuleName, serviceName, "kitex_gen", namespace))},
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
				Name: ast.NewIdent(strings.Title(function.Name)),
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
					Name: ast.NewIdent(strings.Title(function.Name)),
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

func extractPathAfterCache(fullPath string) string {
	index := strings.Index(fullPath, "cache/")
	if index == -1 {
		return ""
	}
	return fullPath[index:]
}
