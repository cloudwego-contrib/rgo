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
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"

	"github.com/cloudwego-contrib/rgo/pkg/common/consts"
	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
	"github.com/cloudwego-contrib/rgo/pkg/db"
	"github.com/cloudwego-contrib/rgo/pkg/tpl"
	"github.com/cloudwego-contrib/rgo/pkg/transformer"
)

func ParseServer(filePaths []string, arg *cmd.Argument) error {
	// traverse current package's file
	for i, fileP := range filePaths {
		// parse go file to ast structure
		fSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fSet, fileP, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse ast file %v failed, err: %v", fileP, err)
		}

		// find main function and parse `rgo.Run(args)` flag
		for _, decl := range astFile.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				// find main function
				if astFile.Name.Name == consts.Main && funcDecl.Name.Name == consts.Main {
					callExpr, err := checkRgoRun(funcDecl, astFile)
					if err != nil {
						return err
					}
					if callExpr != nil {
						handlerDepenMap := make(map[string]*handlerDepenInfo, 20) // key: function name
						tf := transformer.NewThriftFile()                         // thrift file generate structure

						// parse rgo.Run(args)'s args, transformer info to tf, add handlerDepenMap for handler.go
						if err = parseRgoRunArgs(callExpr, tf, astFile, arg, handlerDepenMap); err != nil {
							return err
						}

						srvGenDir := filepath.Join(arg.TempDir, "server", tf.Name)

						// generate thrift idl content
						idlPath := filepath.Join(srvGenDir, strings.ReplaceAll(tf.Namespace, ".", "_")+".thrift")
						if err = os.MkdirAll(filepath.Dir(idlPath), 0o777); err != nil {
							return err
						}
						idlContent, err := tf.Generate(idlPath)
						if err != nil {
							return err
						}

						// execute kitex command
						if err = executeKitexGen(arg, srvGenDir, idlPath, tf.Namespace, true); err != nil {
							return err
						}

						// replace main body
						if err = replaceMainBody(funcDecl, arg, tf, fSet, astFile); err != nil {
							return err
						}
						// replace origin main.go path
						filePaths[i] = filepath.Join(arg.TempDir, "server", "main.go")

						// replace handler.go body
						if err = replaceHandlerBody(handlerDepenMap, filepath.Join(srvGenDir, "handler.go")); err != nil {
							return err
						}

						// handle compile dependence
						if err = handleCompileDependence(tpl.BaseSrvDependence, arg, true); err != nil {
							return err
						}

						// check if it is an existing server version
						sms := make([]*db.ServerManage, 0, 10)
						if err = db.MGetSMByServiceName(context.Background(), tf.Namespace, &sms); err != nil {
							return err
						}
						flag := 0
						for i := len(sms) - 1; i >= 0; i-- {
							if utils.Hash(idlContent) == utils.Hash(sms[i].IdlContent) {
								flag = 1
								break
							}
						}
						if flag == 0 {
							version := ""
							if len(sms) == 0 {
								version = "v1"
							} else {
								n, _ := strconv.Atoi(string(sms[len(sms)-1].ServerVersion[1]))
								version = "v" + strconv.Itoa(n+1)
							}
							// store to db
							if err = db.CreateSM(context.Background(), &db.ServerManage{
								ServiceName:   tf.Namespace,
								IdlContent:    idlContent,
								ServerVersion: version,
							}); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

type handlerDepenInfo struct {
	stReqImportPath string
	funcImptPath    string
	f               *astFuncInfo
}

func replaceHandlerBody(handlerDepenMap map[string]*handlerDepenInfo, handlerPath string) error {
	fSet := token.NewFileSet()
	handlerFile, err := parser.ParseFile(fSet, handlerPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, decl := range handlerFile.Decls {
		if fun, ok := decl.(*ast.FuncDecl); ok {
			sec := fun.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
			secX := sec.X.(*ast.Ident)
			respType := "*" + secX.Name + "." + sec.Sel.Name
			info := handlerDepenMap[fun.Name.Name]
			bodySt := &tpl.SrvHandlerBody{
				PackageName:  filepath.Base(info.funcImptPath),
				FuncName:     fun.Name.Name,
				NeedReqType:  info.f.params[1].pType,
				RealRespType: respType,
			}
			data, err := bodySt.Render()
			if err != nil {
				return err
			}
			utils.ReplaceFuncBody(fun, data)
			utils.AppendImports(fSet, handlerFile, []string{info.stReqImportPath, info.funcImptPath})
			tpl.AddDependence(tpl.BaseSrvDependence, []string{info.stReqImportPath, info.funcImptPath})
		}
	}
	utils.AppendImports(fSet, handlerFile, []string{"unsafe"})

	handlerFile.Name.Name = "handler"

	var buf bytes.Buffer
	if err = format.Node(&buf, fSet, handlerFile); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(filepath.Dir(handlerPath), "handler"), 0o777); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(filepath.Dir(handlerPath), "handler", "handler.go"), buf.Bytes(), 0o777); err != nil {
		return err
	}

	return nil
}

func checkRgoRun(funcDecl *ast.FuncDecl, astFile *ast.File) (*ast.CallExpr, error) {
	for _, stmt := range funcDecl.Body.List {
		if stmtExpr, ok := stmt.(*ast.ExprStmt); ok {
			if stmtExprX, ok := stmtExpr.X.(*ast.CallExpr); ok {
				if stmtExprXFun, ok := stmtExprX.Fun.(*ast.SelectorExpr); ok {
					if stmtExprXFunX, ok := stmtExprXFun.X.(*ast.Ident); ok {
						// find rgo.Run(args) flag
						if stmtExprXFunX.Name == "rgo" && stmtExprXFun.Sel.Name == "Run" {
							// check if the import of rgo.Run(args)'s is correct
							rgoPath, err := utils.GetImportPath("rgo", astFile)
							if err != nil {
								return nil, fmt.Errorf("get rgo import path failed, err: %v", err)
							}
							if rgoPath != consts.RgoImportPath {
								return nil, fmt.Errorf("rgo import path expected %v, extually %v", consts.RgoImportPath, rgoPath)
							}
							return stmtExprX, nil
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func parseRgoRunArgs(callExpr *ast.CallExpr, tf *transformer.ThriftFile, astFile *ast.File,
	cmdArg *cmd.Argument, handlerDepenMap map[string]*handlerDepenInfo) error {
	if len(callExpr.Args) < 2 {
		return errors.New("rgo.Run() function needs al least two arguments: serviceName and one server function")
	}

	// Note: only support the form of rgo.Run("p.s.m", func1, func2, func3), other forms cannot be resolved.
	for i, arg := range callExpr.Args {
		// i0: serviceName
		if i == 0 {
			// check serviceName type
			srvNameType, ok := arg.(*ast.BasicLit)
			if !ok {
				return errors.New("the first param of rgo.Run() must be *ast.BasicLit type")
			}
			if srvNameType.Kind != token.STRING {
				return errors.New("the first param of rgo.Run() must be string type")
			}
			// assign serviceName to tf
			tf.Name = strings.ReplaceAll(srvNameType.Value[1:len(srvNameType.Value)-1], ".", "_")
			tf.Namespace = srvNameType.Value[1 : len(srvNameType.Value)-1]
			tf.Service.Name = tf.Name
		} else {
			// i1-: server function
			// Note: server function can not be written in main package because of handler.go's call
			// check server function type
			argSec, ok := arg.(*ast.SelectorExpr)
			if !ok {
				return errors.New("the server function param of rgo.Run() must be *ast.SelectorExpr type")
			}
			argSecX, ok := argSec.X.(*ast.Ident)
			if !ok {
				return errors.New("the server function param's X of rgo.Run() must be *ast.Ident")
			}
			pkgName := argSecX.Name
			funcName := argSec.Sel.Name

			// get target function's ast type
			funcImptPath, err := utils.GetImportPath(pkgName, astFile)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(cmdArg.GoMod, funcImptPath)
			if err != nil {
				return fmt.Errorf("file path %v %v rel failed, err: %v", cmdArg.GoMod, funcImptPath, err)
			}
			funcDir := filepath.Join(cmdArg.GoModPath, rel)
			funcFile, err := utils.GetAstFileByFuncName(funcDir, funcName)
			if err != nil {
				return err
			}
			funcDecl := utils.GetFunction(funcName, funcFile)
			if funcDecl == nil {
				return fmt.Errorf("can not find %v", funcName)
			}

			// parse server function to tf
			funcInfo, stReqImportPath, err := parseServerFunc(funcDecl, tf, cmdArg, funcFile, funcDir)
			if err != nil {
				return err
			}

			// add handlerDepenMap for handler.go
			handlerDepenMap[funcInfo.funcName] = &handlerDepenInfo{
				stReqImportPath: stReqImportPath,
				funcImptPath:    funcImptPath,
				f:               funcInfo,
			}
		}
	}

	return nil
}

func parseServerFunc(funcDecl *ast.FuncDecl, tf *transformer.ThriftFile, arg *cmd.Argument,
	astFile *ast.File, fileDir string) (*astFuncInfo, string, error) {
	// check grammar errors and parse function name, params info and returns info to AstFuncInfo struct
	funcInfo, err := checkAndParseFunc(funcDecl, astFile.Name.Name)
	if err != nil {
		return nil, "", fmt.Errorf("check and parse function failed, err: %v", err)
	}
	reqParam := funcInfo.params[1]
	respParam := funcInfo.returns[0]

	// parse request structure
	// stReqImportPath is used to build handler dependence
	stReq, stReqImportPath, err := getParsedStruct(&reqParam, arg, astFile, fileDir)
	if err != nil {
		return nil, "", fmt.Errorf("parse struct %v failed, err: %v", reqParam.pType, err)
	}
	tf.Structs = append(tf.Structs, stReq)
	// parse response structure
	stResp, _, err := getParsedStruct(&respParam, arg, astFile, fileDir)
	if err != nil {
		return nil, "", fmt.Errorf("parse struct %v failed, err: %v", reqParam.pType, err)
	}
	tf.Structs = append(tf.Structs, stResp)

	// parse server function
	tf.Service.Functions = append(tf.Service.Functions, &transformer.Function{
		Name:     funcInfo.funcName,
		ReqName:  funcInfo.params[1].name,
		ReqType:  strings.Split(reqParam.pType, ".")[1],
		RespType: strings.Split(respParam.pType, ".")[1],
	})

	return funcInfo, stReqImportPath, nil
}

func replaceMainBody(funcDecl *ast.FuncDecl, cmdArg *cmd.Argument, tf *transformer.ThriftFile,
	fSet *token.FileSet, astFile *ast.File) error {
	srvGenDir := filepath.Join(cmdArg.TempDir, "server", tf.Name)
	idlSrvName := tf.Service.Name
	idlSrvNameLower := strings.ToLower(strings.Replace(idlSrvName, "_", "", -1))
	np := strings.ReplaceAll(tf.Namespace, ".", string(filepath.Separator))

	// find Impl structure in handler.go
	f := token.NewFileSet()
	handlerFile, err := parser.ParseFile(f, filepath.Join(srvGenDir, "handler.go"), nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse ast file %v failed, err: %v", handlerFile, err)
	}
	implName := ""
	ast.Inspect(handlerFile, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeSpec); ok && strings.Contains(t.Name.Name, "Impl") {
			implName = t.Name.Name
			return false
		}
		return true
	})
	if implName == "" {
		return errors.New("can find Impl struct in handler.go")
	}

	// joint dependence
	p, err := filepath.Rel(cmdArg.GoModPath, srvGenDir)
	if err != nil {
		return fmt.Errorf("file path %v %v rel failed, err: %v", cmdArg.GoModPath, srvGenDir, err)
	}
	depenPath := strings.ReplaceAll(filepath.Join(cmdArg.GoMod, p, "kitex_gen", np, idlSrvNameLower), "\\", "/")
	depenPathHandler := strings.ReplaceAll(filepath.Join(cmdArg.GoMod, p, "handler"), "\\", "/")

	// render SrvMainBody
	pkg := filepath.Base(depenPath)
	body := &tpl.SrvMainBody{
		PackageName: pkg,
		ServerImpl:  filepath.Base(depenPathHandler) + "." + implName,
	}
	data, err := body.Render()
	if err != nil {
		return err
	}

	// add body to main.go
	funcDecl.Body.List = append(funcDecl.Body.List, &ast.ExprStmt{
		X: &ast.BasicLit{
			Kind:  token.STRING,
			Value: data,
		},
	})

	// add relevant imports
	utils.AppendImports(fSet, astFile, []string{depenPath, depenPathHandler, "net", "github.com/cloudwego/kitex/server"})

	// add relevant dependence to tpl.BaseSrvDependence
	tpl.BaseSrvDependence[depenPath] = struct{}{}
	tpl.BaseSrvDependence[depenPathHandler] = struct{}{}

	// update the main.go file content and replace the original path
	newPath := filepath.Join(cmdArg.TempDir, "server", "main.go")
	var buf bytes.Buffer
	if err = format.Node(&buf, fSet, astFile); err != nil {
		return err
	}
	if err = os.WriteFile(newPath, buf.Bytes(), 0o777); err != nil {
		return err
	}

	return nil
}
