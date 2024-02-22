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
	"github.com/cloudwego-contrib/rgo/pkg/transformer"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"

	"github.com/cloudwego-contrib/rgo/pkg/common/consts"
	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
	"github.com/cloudwego-contrib/rgo/pkg/db"
	"github.com/cloudwego-contrib/rgo/pkg/tpl"
	thriftgo_parser "github.com/cloudwego/thriftgo/parser"
)

// ParseClient parses function call with `rgo:client` flag and inject code for rpc calls at compile time.
func ParseClient(filePaths []string, arg *cmd.Argument) error {
	for i, fileP := range filePaths {
		// parse go file to ast structure
		fSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fSet, fileP, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse ast file %v failed, err: %v", fileP, err)
		}

		for _, decl := range astFile.Decls {
			var (
				funcDec      *ast.FuncDecl
				sliceComment []string
				idlPath      string
				psm          string
			)

			// find `rgo:` flag and record
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Doc != nil && strings.Contains(funcDecl.Doc.Text(), consts.RgoClient) {
					// check and parse function comments and extract service_name, version
					sliceComment, err = checkAndParseComment(funcDecl.Doc.Text())
					if err != nil {
						return err
					}
					funcDec = funcDecl
				}
			}

			if funcDec != nil {
				// check grammar errors and parse function name, params info and returns info to AstFuncInfo struct
				funcInfo, err := checkAndParseFunc(funcDec, astFile.Name.Name)
				if err != nil {
					return fmt.Errorf("check and parse function failed, err: %v", err)
				}

				// get server version and psm based on comments
				version := ""
				var sm *db.ServerManage
				if strings.Contains(sliceComment[1], "@") {
					// user specifies server version
					version = strings.Split(sliceComment[1], "@")[1]
					psm = strings.Split(sliceComment[1], "@")[0]
				} else {
					// user does not specify server version, default latest
					// get latest version by service name from db
					sm = new(db.ServerManage)
					if err = db.GetLastSMByServiceName(context.Background(), sliceComment[1], sm); err != nil {
						return fmt.Errorf("get latest version from db failed, err: %v", err)
					}
					version = sm.ServerVersion
					psm = sliceComment[1]
				}
				if sm == nil {
					// called when the user specifies a version
					sm = new(db.ServerManage)
					if err = db.GetSMByServiceNameVersion(context.Background(), psm, version, sm); err != nil {
						return fmt.Errorf("get server management info by service name and "+
							"version failed, err: %v", err)
					}
				}
				serviceName := strings.ReplaceAll(psm, ".", "_")
				cliGenDir := filepath.Join(arg.TempDir, "client", serviceName)

				// generate cliGenDir
				isExist, err := utils.PathExist(cliGenDir)
				if err != nil {
					return fmt.Errorf("judge path %v whether exists failed, err: %v", idlPath, err)
				}
				if !isExist {
					if err = os.MkdirAll(cliGenDir, 0o777); err != nil {
						return fmt.Errorf("mkdir dir %v failed, err: %v", cliGenDir, err)
					}
				}

				// write idl to disk
				idlPath = filepath.Join(cliGenDir, serviceName+".thrift")
				isExist, err = utils.PathExist(idlPath)
				if err != nil {
					return fmt.Errorf("judge path %v whether exists failed, err: %v", idlPath, err)
				}
				if !isExist {
					if err = os.WriteFile(idlPath, []byte(sm.IdlContent), 0o777); err != nil {
						return fmt.Errorf("write file %v failed, err: %v", idlPath, err)
					}
				}

				reqParam := funcInfo.params[1]
				respParam := funcInfo.returns[0]
				// parse request structure and check if the versions are consistent
				stReq, _, err := getParsedStruct(&reqParam, arg, astFile, filepath.Dir(fileP))
				if err != nil {
					return fmt.Errorf("parse struct %v failed, err: %v", reqParam.pType, err)
				}
				if checkStructConsist(stReq, sm.IdlContent) == false {
					return fmt.Errorf("client struct inconsistent with server")
				}
				// parse response structure and check if the versions are consistent
				stResp, _, err := getParsedStruct(&respParam, arg, astFile, filepath.Dir(fileP))
				if err != nil {
					return fmt.Errorf("parse struct %v failed, err: %v", reqParam.pType, err)
				}
				if checkStructConsist(stResp, sm.IdlContent) == false {
					return fmt.Errorf("client struct inconsistent with server")
				}

				// parse thrift idl
				astTree, err := thriftgo_parser.ParseFile(idlPath, []string{"."}, false)
				if err != nil {
					return fmt.Errorf("parse thrift idl %v failed, err: %v", idlPath, err)
				}
				if astTree.Services == nil {
					return errors.New("thrift file does not define service")
				}
				// find if the target function exists
				// generate one service by default
				hasTargetFunc := false
				for _, fun := range astTree.Services[0].Functions {
					if fun.Name == funcInfo.funcName {
						idlReqType := fun.Arguments[0].Type.Name
						idlRespType := fun.FunctionType.Name

						var reqSt *thriftgo_parser.StructLike
						var respSt *thriftgo_parser.StructLike
						for _, st := range astTree.Structs {
							if st.Name == idlReqType {
								reqSt = st
							}
							if st.Name == idlRespType {
								respSt = st
							}
						}
						if reqSt == nil || respSt == nil {
							return errors.New("can not find remote repository structure")
						}
						hasTargetFunc = true
					}
				}
				if hasTargetFunc == false {
					return errors.New("can not find specified function")
				}

				// get client dependency import paths
				depend, err := getCliDependImportPath(arg, astTree, cliGenDir)
				if err != nil {
					return err
				}
				// add imports to dependence map
				tpl.AddDependence(tpl.BaseClientDependence, []string{
					depend.srvNameLowerPath,
					depend.namespacePath, depend.kitexGenPath,
				})

				isExist, err = utils.PathExist(filepath.Join(cliGenDir, "kitex_gen"))
				if err != nil {
					return fmt.Errorf("judge whether dir: %v exists failed, err: %v",
						filepath.Join(cliGenDir, "kitex_gen"), err)
				}
				if !isExist {
					// execute kitex command
					if err = executeKitexGen(arg, cliGenDir, idlPath, psm, false); err != nil {
						return err
					}

					// generate client.go
					if err = genInitClient(depend, astTree, cliGenDir, psm); err != nil {
						return err
					}
				}
				// replace function body
				if err := replaceClientFuncBody(depend, astTree, funcInfo, fSet, astFile, funcDec); err != nil {
					return err
				}

				// update the file content and replace the original path
				newPath := filepath.Join(cliGenDir, filepath.Base(fileP))
				var buf bytes.Buffer
				if err = format.Node(&buf, fSet, astFile); err != nil {
					return fmt.Errorf("format ast file content failed, err: %v", err)
				}
				if err = os.WriteFile(newPath, buf.Bytes(), 0o777); err != nil {
					return fmt.Errorf("write file %v failed, err: %v", newPath, err)
				}
				filePaths[i] = newPath
			}
		}
	}

	// handle compile dependence
	// BaseClientDependence init length = 4
	if len(tpl.BaseClientDependence) != 4 {
		if err := handleCompileDependence(tpl.BaseClientDependence, arg, false); err != nil {
			return err
		}
	}
	return nil
}

func checkStructConsist(st *transformer.Struct, idlContent string) bool {
	if !strings.Contains(idlContent, st.GenerateSingle()) {
		return false
	}
	for _, field := range st.StructFields {
		if len(field.RelevantStructs) != 0 {
			for _, subSt := range field.RelevantStructs {
				if !checkStructConsist(subSt, idlContent) {
					return false
				}
			}
		}
		if len(field.RelevantEnums) != 0 {
			for _, subEnum := range field.RelevantEnums {
				if !strings.Contains(idlContent, subEnum.Generate()) {
					return false
				}
			}
		}
	}
	return true
}

type cliDependImportPath struct {
	idlSrvNameLower  string
	namespacePath    string
	srvNameLowerPath string
	kitexGenPath     string
}

func getCliDependImportPath(arg *cmd.Argument, astTree *thriftgo_parser.Thrift, cliGenDir string) (*cliDependImportPath, error) {
	depend := &cliDependImportPath{}
	idlSrvName := astTree.Services[0].Name
	depend.idlSrvNameLower = strings.ToLower(strings.Replace(idlSrvName, "_", "", -1))
	np := strings.ReplaceAll(astTree.Namespaces[0].Name, ".", string(filepath.Separator))
	p, err := filepath.Rel(arg.GoModPath, cliGenDir)
	if err != nil {
		return nil, fmt.Errorf("relative path %v %v failed, err: %v", arg.GoModPath, cliGenDir, err)
	}
	depend.namespacePath = strings.ReplaceAll(filepath.Join(arg.GoMod, p, "kitex_gen", np), "\\", "/")
	depend.srvNameLowerPath = strings.ReplaceAll(filepath.Join(arg.GoMod, p, "kitex_gen", np,
		depend.idlSrvNameLower), "\\", "/")
	depend.kitexGenPath = strings.ReplaceAll(filepath.Join(arg.GoMod, p, "kitex_gen"), "\\", "/")
	return depend, nil
}

func executeKitexGen(arg *cmd.Argument, genDir, idlPath, psm string, isServer bool) error {
	// change work dir to genDir
	if err := os.Chdir(genDir); err != nil {
		return fmt.Errorf("changing work directory failed, err: %v", err)
	}
	// execute kitex command to generate idl content
	if err := utils.KitexGen(isServer, arg.GoMod, idlPath, psm); err != nil {
		return fmt.Errorf("exectuate kitex command failed, err: %v", err)
	}
	// change work dir to current dir
	if err := os.Chdir(arg.Cwd); err != nil {
		return fmt.Errorf("changing work directory failed, err: %v", err)
	}
	// execute go mod tidy
	if err := utils.GoModTidy(); err != nil {
		return fmt.Errorf("execute go mod tidy failed, err: %v", err)
	}
	return nil
}

func genInitClient(depend *cliDependImportPath, astTree *thriftgo_parser.Thrift, cliGenDir, psm string) error {
	cliTpl := &tpl.ClientGoTemplate{
		Imports: map[string]struct{}{
			depend.srvNameLowerPath: {},
		},
		IdlSrvNameLower: depend.idlSrvNameLower,
		Psm:             psm,
		IdlFileName:     strings.ToUpper(astTree.Services[0].Name),
	}
	data, err := cliTpl.Render()
	if err != nil {
		return fmt.Errorf("render client go tpl failed, err: %v", err)
	}
	if err = os.WriteFile(filepath.Join(cliGenDir, "kitex_gen", "client.go"), []byte(data), 0o777); err != nil {
		return fmt.Errorf("write file %v failed, err: %v", filepath.Join(cliGenDir, "kitex_gen", "client.go"), err)
	}
	return nil
}

func replaceClientFuncBody(depend *cliDependImportPath, astTree *thriftgo_parser.Thrift, funcInfo *astFuncInfo,
	fSet *token.FileSet, astFile *ast.File, funcDec *ast.FuncDecl) error {
	respType := ""
	if funcInfo.returns[0].isSamePkg {
		respType = "*" + strings.Split(funcInfo.returns[0].pType, ".")[1]
	} else {
		respType = funcInfo.returns[0].pType
	}
	funcBody := tpl.ClientFuncBody{
		ClientName:      strings.ToUpper(astTree.Services[0].Name) + "Client",
		FuncName:        funcInfo.funcName,
		FirstParamName:  funcInfo.params[0].name,
		SecondParamName: funcInfo.params[1].name,
		RealReqType:     "*" + filepath.Base(depend.namespacePath) + "." + strings.Split(funcInfo.params[1].pType, ".")[1],
		NeedRespType:    respType,
	}
	body, err := funcBody.Render()
	if err != nil {
		return fmt.Errorf("render client funciton body failed, err: %v", err)
	}
	utils.AppendImports(fSet, astFile, []string{depend.namespacePath, depend.kitexGenPath, "unsafe", "github.com/cloudwego/kitex/client/callopt"})
	utils.ReplaceFuncBody(funcDec, body)
	return nil
}

type compileDependInfo struct {
	genGoPath  string
	genExePath string
	genTxtPath string
}

func getCompileDependInfo(arg *cmd.Argument, isServer bool) (*compileDependInfo, error) {
	info := &compileDependInfo{}
	basePath := ""
	if isServer {
		basePath = filepath.Join(arg.TempDir, "depen", "server")
	} else {
		basePath = filepath.Join(arg.TempDir, "depen", "client")
	}

	isExist, err := utils.PathExist(basePath)
	if err != nil {
		return nil, fmt.Errorf("judge path %v whether exists failed, err: %v", basePath, err)
	}
	if !isExist {
		if err = os.MkdirAll(basePath, 0o777); err != nil {
			return nil, fmt.Errorf("mkdir dir %v failed, err: %v", basePath, err)
		}
	}

	info.genGoPath = filepath.Join(basePath, "depency.go")
	info.genExePath = filepath.Join(basePath, "depency")
	info.genTxtPath = filepath.Join(basePath, "depency.txt")
	return info, nil
}

func handleCompileDependence(dependenceMap map[string]struct{}, arg *cmd.Argument, isServer bool) error {
	data, err := tpl.RenderAnonymousDependenceTpl(dependenceMap)
	if err != nil {
		return fmt.Errorf("render anonymous dependence tpl failed, err: %v", err)
	}

	info, err := getCompileDependInfo(arg, isServer)
	if err != nil {
		return err
	}

	if err = utils.BuildDependence(info.genGoPath, info.genExePath, info.genTxtPath, data); err != nil {
		return fmt.Errorf("build compile phase imports dependence failed, err: %v", err)
	}

	// merge import config
	if isServer {
		if err = utils.MergeImportCfg(info.genTxtPath, arg.ServerCfg, "importcfg", true); err != nil {
			return fmt.Errorf("merge compile phase imports dependence failed, err: %v", err)
		}
	} else {
		if err = utils.MergeImportCfg(info.genTxtPath, arg.ClientCfg, "importcfg", true); err != nil {
			return fmt.Errorf("merge compile phase imports dependence failed, err: %v", err)
		}
	}

	return nil
}
