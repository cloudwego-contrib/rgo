package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego-contrib/rgo/utils"
	"github.com/cloudwego/thriftgo/parser"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerateRGOCode(idlRepoPath, idlPath, rgoRepoPath string) error {
	fileType := filepath.Ext(idlPath)
	idlRepoPath = strings.ReplaceAll(idlRepoPath, "@", "/")

	switch fileType {
	case ".thrift":
		err := generateThriftCode(idlRepoPath, idlPath, rgoRepoPath)
		if err != nil {
			return err
		}

		return generateClientCode(idlRepoPath, idlPath, rgoRepoPath)
	case ".proto":
		return nil
	default:
		return errors.New("unsupported idl file")
	}
}

func generateClientCode(idlRepoPath, idlPath, rgoRepoPath string) error {
	thriftFile, err := parseIDLFile(idlPath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()

	namespace, f, err := buildThriftAstFile(thriftFile)
	if err != nil {
		return err
	}

	exist, err := utils.FileExistsInPath(rgoRepoPath, "go.mod")
	if err != nil {
		return err
	}

	if !exist {
		err = initGoMod(consts.RGOModuleName, rgoRepoPath)
		if err != nil {
			return err
		}
	}

	outputDir := filepath.Join(rgoRepoPath, consts.RGOGenCodePath, idlRepoPath, namespace)

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

	return nil
}

func buildThriftAstFile(thrift *parser.Thrift) (string, *ast.File, error) {
	var namespace string

	for _, v := range thrift.Namespaces {
		if v.Language == "go" {
			namespace = strings.ToLower(v.GetName())
		}
	}

	if namespace == "" {
		return "", nil, errors.New("no go namespace found")
	}

	f := &ast.File{
		Name: ast.NewIdent(namespace),
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
