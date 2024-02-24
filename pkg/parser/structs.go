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
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/cmd"
	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
	"github.com/cloudwego-contrib/rgo/pkg/transformer"
)

// getParsedStruct is used to parse param.pType to *transformer.Struct.
// There are three situations: within the local package, outside the local package, and in the remote repo.
func getParsedStruct(param *param, arg *cmd.Argument, astFile *ast.File, fileDir string) (*transformer.Struct, string, error) {
	if param.isSamePkg {
		// within the local package
		f, astSt, err := utils.GetAstFileByStructName(fileDir, strings.Split(param.pType, ".")[1])
		if err != nil {
			return nil, "", err
		}
		st, err := transformer.ConvertStruct(strings.Split(param.pType, ".")[1], arg.GoModPath, arg.GoMod, fileDir, astSt, f)
		if err != nil {
			return nil, "", err
		}
		p, err := filepath.Rel(arg.GoModPath, fileDir)
		if err != nil {
			return nil, "", fmt.Errorf("file path %v %v rel failed, err: %v", arg.GoModPath, fileDir, err)
		}
		return st, strings.ReplaceAll(filepath.Join(arg.GoMod, p), "\\", "/"), nil
	} else {
		imptPath, err := utils.GetImportPath(strings.Split(param.pType, ".")[0][1:], astFile)
		if err != nil {
			return nil, "", err
		}

		p, _ := filepath.Rel(arg.GoMod, imptPath)
		isExist, err := utils.PathExist(filepath.Join(arg.GoModPath, p))
		if err != nil {
			return nil, "", err
		}
		if !isExist {
			// in the remote repo
			stInfo, err := getRemoteStructInfo(strings.Split(param.pType, ".")[0][1:], arg.GoModPath, astFile)
			if err != nil {
				return nil, "", err
			}
			rel, err := filepath.Rel(stInfo.repoImportPath, stInfo.stImportPath)
			if err != nil {
				return nil, "", err
			}
			stPath := filepath.Join(stInfo.repoPath, rel)
			remoteF, astSt, err := utils.GetAstFileByStructName(stPath, strings.Split(param.pType, ".")[1])
			if err != nil {
				return nil, "", err
			}
			st, err := transformer.ConvertStruct(strings.Split(param.pType, ".")[1], stInfo.repoPath, stInfo.repoImportPath, stPath, astSt, remoteF)
			if err != nil {
				return nil, "", err
			}
			return st, stInfo.stImportPath, nil
		} else {
			// outside the local package
			f, astSt, err := utils.GetAstFileByStructName(filepath.Join(arg.GoModPath, p), strings.Split(param.pType, ".")[1])
			if err != nil {
				return nil, "", err
			}

			st, err := transformer.ConvertStruct(strings.Split(param.pType, ".")[1], arg.GoModPath, arg.GoMod, filepath.Join(arg.GoModPath, p), astSt, f)
			if err != nil {
				return nil, "", err
			}
			return st, imptPath, nil
		}
	}
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
