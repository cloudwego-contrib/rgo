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
	"path"
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

// ParseServer parses function call with `rgo:server` flag and inject code for rpc server at compile time.
func ParseServer(filePaths []string, arg *cmd.Argument) error {
	var (
		srvGenDir string
		recordPsm string
	)
	funcMap := make(map[string]*funcDepenInfo, 20) // key: function name
	tf := transformer.NewThriftFile()

	for i, fileP := range filePaths {
		// parse go file to ast structure
		fSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fSet, fileP, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse ast file %v failed, err: %v", fileP, err)
		}

		// find `rgo:server` and `rgo.Run()` flag
		hasRgoTag := false
		for _, decl := range astFile.Decls {
			var funcDec *ast.FuncDecl
			sliceComment := make([]string, 0, 5)
			// rgo.Run() logic
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Doc != nil && strings.Contains(funcDecl.Doc.Text(), consts.RgoServer) {
					// parse function comment and extract p.s.m, dns, version
					sliceComment, err = checkAndParseComment(funcDecl.Doc.Text(), true)
					if err != nil {
						return err
					}
					funcDec = funcDecl
				}
				if astFile.Name.Name == "main" && funcDecl.Name.Name == "main" {
					hasRgoTag, err = handleMainGo(arg, funcDecl, fSet, astFile)
					if err != nil {
						return err
					}
				}
			}

			if hasRgoTag {
				// handle compile dependence
				if err = handleCompileDependence(tpl.BaseSrvMainDependence, arg, true); err != nil {
					return err
				}

				// update the main.go file content and replace the original path
				newPath := filepath.Join(arg.TempDir, "server", "main.go")
				var buf bytes.Buffer
				if err = format.Node(&buf, fSet, astFile); err != nil {
					return err
				}
				if err = os.WriteFile(newPath, buf.Bytes(), 0o777); err != nil {
					return err
				}
				filePaths[i] = newPath
			}

			if funcDec != nil {
				// check grammar errors and parse function name, params info and returns info to AstFuncInfo struct
				funcInfo, err := checkAndParseFunc(funcDec)
				if err != nil {
					return fmt.Errorf("check and parse function failed, err: %v", err)
				}

				psm := sliceComment[4]
				if recordPsm == "" {
					recordPsm = psm
				}
				if recordPsm != psm {
					return fmt.Errorf("multiple services cannot be specified, existed: %v, new: %v", recordPsm, psm)
				}
				serviceName := strings.ReplaceAll(psm, ".", "_")
				srvGenDir = path.Join(arg.TempDir, "server", serviceName)
				isExist, err := utils.PathExist(srvGenDir)
				if err != nil {
					return fmt.Errorf("judge whether dir: %v exists failed", err)
				}
				if isExist {
					return fmt.Errorf("need to define server functions in one package")
				}

				// record tf structure
				tf.Name = serviceName
				tf.Namespace = psm
				// convert remote repository request and response
				// request
				stReq, stReqImportPath, err := getParsedStruct(&funcInfo.Params[1], arg, astFile)
				if err != nil {
					return err
				}
				tf.Structs = append(tf.Structs, stReq)
				// response
				stResp, _, err := getParsedStruct(&funcInfo.Returns[0], arg, astFile)
				if err != nil {
					return err
				}
				tf.Structs = append(tf.Structs, stResp)
				// todo optimize tf.Service.Name
				// record thrift service and it's function
				tf.Service.Name = strings.Title(serviceName)
				tf.Service.Functions = append(tf.Service.Functions, &transformer.Function{
					Name:     funcInfo.FuncName,
					ReqName:  funcInfo.Params[1].Name,
					ReqType:  strings.Split(funcInfo.Params[1].Type, ".")[1],
					RespType: strings.Split(funcInfo.Returns[0].Type, ".")[1],
				})

				p, err := filepath.Rel(arg.GoModPath, filepath.Dir(fileP))
				funPath := strings.ReplaceAll(filepath.Join(arg.GoMod, p), "\\", "/")
				if err != nil {
					return err
				}
				// add dependence for `tpl.BaseSrvHandlerDependence`
				tpl.AddDependence(tpl.BaseSrvHandlerDependence, []string{stReqImportPath, funPath})
				funcMap[funcInfo.FuncName] = &funcDepenInfo{
					depen: []string{stReqImportPath, funPath},
					f:     funcInfo,
				}
			}
		}
	}

	if len(funcMap) != 0 {
		// generate thrift idl content
		idlPath := filepath.Join(srvGenDir, strings.ReplaceAll(recordPsm, ".", "_")+".thrift")
		if err := os.MkdirAll(filepath.Dir(idlPath), 0o777); err != nil {
			return err
		}
		idlContent, err := tf.Generate(idlPath)
		if err != nil {
			return err
		}

		// execute kitex command
		if err = executeKitexGen(arg, srvGenDir, idlPath, recordPsm, true); err != nil {
			return err
		}

		// generate record.txt
		if err = generateRecordTxt(arg, srvGenDir, tf, recordPsm); err != nil {
			return err
		}

		// replace handler.go body
		if err = replaceHandlerBody(funcMap, filepath.Join(srvGenDir, "handler.go")); err != nil {
			return err
		}

		// handle compile dependence
		if err = handleCompileDependence(tpl.BaseSrvHandlerDependence, arg, true); err != nil {
			return err
		}

		// check if it is an existing server version
		sms := make([]*db.ServerManage, 0, 10)
		if err = db.MGetSMByServiceName(context.Background(), recordPsm, sms); err != nil {
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
				ServiceName:   recordPsm,
				IdlContent:    idlContent,
				ServerVersion: version,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

type funcDepenInfo struct {
	depen []string
	f     *AstFuncInfo
}

func replaceHandlerBody(funcMap map[string]*funcDepenInfo, handlerPath string) error {
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
			info := funcMap[fun.Name.Name]
			bodySt := &tpl.SrvHandlerBody{
				PackageName:  filepath.Base(info.depen[1]),
				FuncName:     fun.Name.Name,
				NeedReqType:  info.f.Params[1].Type,
				RealRespType: respType,
			}
			data, err := bodySt.Render()
			if err != nil {
				return err
			}
			utils.ReplaceFuncBody(fun, data)
			utils.AppendImports(fSet, handlerFile, []string{info.depen[0], info.depen[1]})
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

func getParsedStruct(param *Param, arg *cmd.Argument, astFile *ast.File) (*transformer.Struct, string, error) {
	stInfo, err := getRemoteStructInfo(strings.Split(param.Type, ".")[0][1:],
		arg.GoModPath, astFile)
	if err != nil {
		return nil, "", err
	}
	rel, err := filepath.Rel(stInfo.repoImportPath, stInfo.stImportPath)
	if err != nil {
		return nil, "", err
	}
	stPath := filepath.Join(stInfo.repoPath, rel)
	f, err := transformer.GetRemoteStruct(stPath, strings.Split(param.Type, ".")[1])
	if err != nil {
		return nil, "", err
	}
	comment := stInfo.repoImportPath + " " + stInfo.version
	st, err := transformer.ConvertStruct(strings.Split(param.Type, ".")[1], stInfo.repoPath, stInfo.repoImportPath, comment, f)
	if err != nil {
		return nil, "", err
	}
	return st, stInfo.stImportPath, nil
}

type remoteRepoStInfo struct {
	repoPath       string
	version        string
	repoImportPath string
	stImportPath   string
}

func getRemoteStructInfo(pkgName, goModPath string, astFile *ast.File) (*remoteRepoStInfo, error) {
	importValue, err := utils.GetImportPath(pkgName, astFile)
	if err != nil {
		return nil, err
	}

	fileData, err := os.ReadFile(filepath.Join(goModPath, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("read file %v failed, err: %v", filepath.Join(goModPath, "go.mod"), err)
	}
	fileString := string(fileData)

	// find repoImportPath
	repoImportPath := ""
	for i := len(importValue); i >= 0; i-- {
		if strings.Contains(fileString, importValue[0:i]) {
			repoImportPath = importValue[0:i]
			break
		}
	}
	if repoImportPath == "" {
		return nil, errors.New("can not find repo import path")
	}

	if strings.Contains(fileString, repoImportPath) {
		beginIndex := -1
		index := strings.Index(fileString, repoImportPath)
		for i := index; i < len(fileString); i++ {
			if string(fileString[i]) == "v" {
				beginIndex = i
				break
			}
		}
		if beginIndex == -1 {
			return nil, fmt.Errorf("can not find version at the behind of %v", repoImportPath)
		}

		endIndex := -1
		for i := beginIndex; i < len(fileString); i++ {
			if string(fileString[i]) == "\n" || string(fileString[i]) == " " {
				endIndex = i
				break
			}
		}
		if endIndex == -1 {
			return nil, fmt.Errorf("can not find version end flag at the behind of %v", repoImportPath)
		}

		version := fileString[beginIndex:endIndex]
		repoPath, err := getRemoteRepoPath(repoImportPath, version)
		if err != nil {
			return nil, err
		}

		return &remoteRepoStInfo{
			repoPath:       repoPath,
			version:        version,
			repoImportPath: repoImportPath,
			stImportPath:   importValue,
		}, nil
	} else {
		return nil, fmt.Errorf("can not find %v in gomod file", repoImportPath)
	}
}

func getRemoteRepoPath(importPath, version string) (string, error) {
	goPath, err := utils.GetGOPATH()
	if err != nil {
		return "", fmt.Errorf("get gp path failed, err: %v", err)
	}

	repoPath := filepath.Join(goPath, "pkg", "mod", importPath+"@"+version)
	isExist, err := utils.PathExist(repoPath)
	if err != nil {
		return "", fmt.Errorf("judge path exist failed, path: %v, err: %v", repoPath, err)
	}
	if !isExist {
		return "", fmt.Errorf("file %v not exist", repoPath)
	}

	return repoPath, nil
}

func handleMainGo(arg *cmd.Argument, funcDecl *ast.FuncDecl, fSet *token.FileSet, astFile *ast.File) (bool, error) {
	for index, stmt := range funcDecl.Body.List {
		if stmtExpr, ok := stmt.(*ast.ExprStmt); ok {
			if stmtExprX, ok := stmtExpr.X.(*ast.CallExpr); ok {
				if stmtExprXFun, ok := stmtExprX.Fun.(*ast.SelectorExpr); ok {
					if stmtExprXFunX, ok := stmtExprXFun.X.(*ast.Ident); ok {
						if stmtExprXFunX.Name == "rgo" && stmtExprXFun.Sel.Name == "Run" {
							txtData, err := os.ReadFile(filepath.Join(arg.TempDir, "server", "record.txt"))
							if err != nil {
								return false, err
							}
							pkgName := filepath.Base(strings.Split(string(txtData), " ")[0])
							body := &tpl.SrvMainBody{
								PackageName: pkgName,
								ServerImpl:  filepath.Base(strings.Split(string(txtData), " ")[1]) + "." + strings.Split(string(txtData), " ")[2],
							}
							data, err := body.Render()
							if err != nil {
								return false, err
							}
							funcDecl.Body.List[index] = &ast.ExprStmt{
								X: &ast.BasicLit{
									Kind:  token.STRING,
									Value: data,
								},
							}
							utils.AppendImports(fSet, astFile, []string{
								strings.Split(string(txtData), " ")[0],
								strings.Split(string(txtData), " ")[1], "net", "github.com/cloudwego/kitex/server",
							})
							// delete rgo's import
							p, err := utils.GetImportPath("rgo", astFile)
							if err != nil {
								return false, err
							}
							utils.DeleteImport(fSet, astFile, p)
							tpl.BaseSrvMainDependence[strings.Split(string(txtData), " ")[0]] = struct{}{}
							tpl.BaseSrvMainDependence[strings.Split(string(txtData), " ")[1]] = struct{}{}
							return true, nil
						}
					}
				}
			}
		}
	}
	return false, nil
}

// generateRecordTxt is used to generate record.txt for replace main.go's body.
// txt content: depenPath+" "+depenPath1+" "+implName
func generateRecordTxt(arg *cmd.Argument, srvGenDir string, tf *transformer.ThriftFile, recordPsm string) error {
	p, err := filepath.Rel(arg.GoModPath, srvGenDir)
	if err != nil {
		return err
	}
	idlSrvName := tf.Service.Name
	idlSrvNameLower := strings.ToLower(strings.Replace(idlSrvName, "_", "", -1))
	np := strings.ReplaceAll(recordPsm, ".", string(filepath.Separator))
	fSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fSet, filepath.Join(srvGenDir, "handler.go"), nil, parser.ParseComments)
	if err != nil {
		return err
	}
	implName := ""
	ast.Inspect(astFile, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeSpec); ok && strings.Contains(t.Name.Name, "Impl") {
			implName = t.Name.Name
			return false
		}
		return true
	})
	depenPath := strings.ReplaceAll(filepath.Join(arg.GoMod, p, "kitex_gen", np, idlSrvNameLower), "\\", "/")
	depenPath1 := strings.ReplaceAll(filepath.Join(arg.GoMod, p, "handler"), "\\", "/")
	if err = os.WriteFile(filepath.Join(arg.TempDir, "server", "record.txt"), []byte(depenPath+" "+depenPath1+" "+implName), 0o777); err != nil {
		return fmt.Errorf("write file %v failed, err: %v", filepath.Join(arg.TempDir, "server", "record.txt"), err)
	}
	return nil
}
