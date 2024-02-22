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

package transformer

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/common/utils"
)

const (
	list = "list"
	m    = "map"
)

var TypeMapping = map[string]string{
	"bool":    "bool",
	"int":     "i64",
	"int8":    "byte",
	"int16":   "i16",
	"int32":   "i32",
	"int64":   "i64",
	"uint":    "i64",
	"uint8":   "byte",
	"uint16":  "i16",
	"uint32":  "i32",
	"uint64":  "i64",
	"byte":    "byte",
	"float32": "double",
	"float64": "double",
	"string":  "string",
}

type ThriftFile struct {
	Name      string
	Namespace string
	Structs   []*Struct
	Service   *Service

	structMap map[string]struct{}
	enumMap   map[string]struct{}
}

type Struct struct {
	Name         string
	StructFields []*StructField
}

type StructField struct {
	Name            string
	Type            string
	Tag             string
	RelevantStructs []*Struct
	RelevantEnums   []*Enum
}

type Enum struct {
	Name       string
	EnumFields []*EnumField
}

type EnumField struct {
	Name  string
	Value string
}

type Function struct {
	Name     string
	ReqName  string
	ReqType  string
	RespType string
}

type Service struct {
	Name      string
	Functions []*Function
}

func NewThriftFile() *ThriftFile {
	return &ThriftFile{
		Structs: []*Struct{},
		Service: &Service{
			Functions: []*Function{},
		},
		structMap: make(map[string]struct{}),
		enumMap:   make(map[string]struct{}),
	}
}

func (tf *ThriftFile) Generate(genPath string) (string, error) {
	// generate namespace
	data := "namespace go " + tf.Namespace + "\n\n"

	// generate structs
	for _, st := range tf.Structs {
		if _, ok := tf.structMap[st.Name]; !ok {
			data += st.Generate(tf.structMap, tf.enumMap)
			tf.structMap[st.Name] = struct{}{}
		}
	}

	// generate service
	data += "service " + tf.Service.Name + " {\n"
	for _, fun := range tf.Service.Functions {
		data += "\t" + fun.RespType + " " + fun.Name + "(1: " + fun.ReqType + " " + fun.ReqName + ");\n"
	}
	data += "}"

	if err := os.WriteFile(genPath, []byte(data), 0o777); err != nil {
		return "", fmt.Errorf("generate thrift idl %v failed, err: %v", genPath, err)
	}

	return data, nil
}

func (st *Struct) Generate(structMap map[string]struct{}, enumMap map[string]struct{}) string {
	data := "struct " + st.Name + " {\n"
	stDatas := make([]string, 0, 10)
	enumDatas := make([]string, 0, 3)

	if _, ok := structMap[st.Name]; !ok {
		for index, field := range st.StructFields {
			if field.Tag != "" {
				data += "\t" + strconv.Itoa(index+1) + ": " + field.Type + " " + field.Name + " (" + field.Tag + ");\n"
			} else {
				data += "\t" + strconv.Itoa(index+1) + ": " + field.Type + " " + field.Name + "\n"
			}
			if len(field.RelevantStructs) != 0 {
				for _, subSt := range field.RelevantStructs {
					if _, ok = structMap[field.Name]; !ok {
						stDatas = append(stDatas, subSt.Generate(structMap, enumMap))
						structMap[subSt.Name] = struct{}{}
					}
				}
			}
			if len(field.RelevantEnums) != 0 {
				for _, subEnum := range field.RelevantEnums {
					if _, ok = enumMap[field.Name]; !ok {
						enumDatas = append(enumDatas, subEnum.Generate())
						enumMap[subEnum.Name] = struct{}{}
					}
				}
			}

		}
		structMap[st.Name] = struct{}{}
	}
	data += "}\n\n"

	for _, stData := range stDatas {
		data += stData
	}
	for _, enumData := range enumDatas {
		data += enumData
	}

	return data
}

func (st *Struct) GenerateSingle() string {
	data := "struct " + st.Name + " {\n"
	for index, field := range st.StructFields {
		if field.Tag != "" {
			data += "\t" + strconv.Itoa(index+1) + ": " + field.Type + " " + field.Name + " (" + field.Tag + ");\n"
		} else {
			data += "\t" + strconv.Itoa(index+1) + ": " + field.Type + " " + field.Name + "\n"
		}
	}
	data += "}\n\n"
	return data
}

func (enum *Enum) Generate() string {
	data := "enum " + enum.Name + " {\n"
	for _, field := range enum.EnumFields {
		data += "\t" + field.Name + " = " + field.Value + ";\n"
	}
	data += "}\n\n"
	return data
}

func ConvertStruct(name, repoPath, repoImport, curPath string, astSt *ast.StructType, astFile *ast.File) (*Struct, error) {
	st := &Struct{
		Name:         name,
		StructFields: []*StructField{},
	}
	for _, field := range astSt.Fields.List {
		stField := &StructField{
			Name: field.Names[0].Name,
			Type: getType(field.Type),
		}
		if field.Tag != nil {
			stField.Tag = field.Tag.Value
		}
		sts, enums, err := getRelevantStructsEnums(field.Type, astFile, repoPath, repoImport, curPath)
		if err != nil {
			return nil, err
		}
		if len(sts) != 0 {
			stField.RelevantStructs = sts
		}
		if len(enums) != 0 {
			stField.RelevantEnums = enums
		}
		st.StructFields = append(st.StructFields, stField)
	}
	return st, nil
}

func getRelevantStructsEnums(expr ast.Expr, astFile *ast.File, repoPath, repoImport, curPath string) (sts []*Struct, enums []*Enum, err error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		if _, ok := TypeMapping[expr.Name]; ok {
			// base type, nothing to do
			return nil, nil, nil
		} else {
			// enum
			enum, err := getDirEnum(curPath, expr.Name)
			if err != nil {
				return nil, nil, err
			}
			if enum != nil {
				return nil, []*Enum{enum}, nil
			} else {
				return nil, nil, fmt.Errorf("can not find enum type: %v", expr.Name)
			}
		}
	case *ast.ArrayType:
		return getRelevantStructsEnums(expr.Elt, astFile, repoPath, repoImport, curPath)
	case *ast.MapType:
		keySt, keyEnum, err := getRelevantStructsEnums(expr.Key, astFile, repoPath, repoImport, curPath)
		if err != nil {
			return nil, nil, err
		}
		vSt, vEnum, err := getRelevantStructsEnums(expr.Value, astFile, repoPath, repoImport, curPath)
		if err != nil {
			return nil, nil, err
		}
		if len(keySt) != 0 {
			sts = append(sts, keySt...)
		}
		if len(vSt) != 0 {
			sts = append(sts, vSt...)
		}
		if len(keyEnum) != 0 {
			enums = append(enums, keyEnum...)
		}
		if len(vEnum) != 0 {
			enums = append(enums, vEnum...)
		}
		return sts, enums, nil
	case *ast.SelectorExpr:
		exprX, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil, nil, errors.New("unsupported struct field type")
		}
		searchPath, err := getSearchPath(exprX.Name, repoImport, repoPath, astFile)
		if err != nil {
			return nil, nil, err
		}
		enum, err := getDirEnum(searchPath, expr.Sel.Name)
		if err != nil {
			return nil, nil, err
		}
		return nil, []*Enum{enum}, nil
	case *ast.StarExpr:
		if exprX, ok := expr.X.(*ast.Ident); ok {
			f, astSt, err := utils.GetAstFileByStructName(curPath, exprX.Name)
			if err != nil {
				return nil, nil, err
			}
			stt, err := ConvertStruct(exprX.Name, repoPath, repoImport, curPath, astSt, f)
			if err != nil {
				return nil, nil, err
			}
			return []*Struct{stt}, nil, nil
		} else if exprSe, ok := expr.X.(*ast.SelectorExpr); ok {
			exprSeX, ok := exprSe.X.(*ast.Ident)
			if !ok {
				return nil, nil, errors.New("unsupported struct field type")
			}
			searchPath, err := getSearchPath(exprSeX.Name, repoImport, repoPath, astFile)
			if err != nil {
				return nil, nil, err
			}
			f, astSt, err := utils.GetAstFileByStructName(searchPath, exprSe.Sel.Name)
			if err != nil {
				return nil, nil, err
			}
			stt, err := ConvertStruct(exprSe.Sel.Name, repoPath, repoImport, searchPath, astSt, f)
			if err != nil {
				return nil, nil, err
			}
			return []*Struct{stt}, nil, nil
		} else {
			return nil, nil, errors.New("unsupported star type")
		}
	}
	return nil, nil, nil
}

func getType(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		v, ok := TypeMapping[expr.Name]
		if ok {
			return v
		} else {
			// enum
			return expr.Name
		}
	case *ast.ArrayType:
		return list + "<" + getType(expr.Elt) + ">"
	case *ast.MapType:
		return m + "<" + getType(expr.Key) + "," + getType(expr.Value) + ">"
	case *ast.SelectorExpr:
		return expr.Sel.Name
	case *ast.StarExpr:
		return getType(expr.X)
	}
	return ""
}

func getDirEnum(searchPath, name string) (*Enum, error) {
	files, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, err
	}

	var enum *Enum
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".go") {
			p := filepath.Join(searchPath, f.Name())
			fSet := token.NewFileSet()
			astFile, err := parser.ParseFile(fSet, p, nil, parser.ParseComments)
			if err != nil {
				return nil, err
			}
			enum, err = getEnum(name, astFile)
			if err != nil {
				return nil, err
			}
			if enum != nil {
				break
			}
		}
	}

	if enum == nil {
		return nil, fmt.Errorf("can not find enum: %v", name)
	}

	return enum, nil
}

func getEnum(name string, astFile *ast.File) (*Enum, error) {
	enum := &Enum{
		Name:       name,
		EnumFields: []*EnumField{},
	}
	var ty *ast.TypeSpec
	t := ""
	ast.Inspect(astFile, func(n ast.Node) bool {
		if spec, ok := n.(*ast.TypeSpec); ok {
			if spec.Name.Name == name {
				if _, ok = spec.Type.(*ast.Ident); ok {
					ty = spec
					t = spec.Name.Name
				}
				return false
			}
		}
		return true
	})

	if ty == nil {
		return nil, nil
	}

	for _, decl := range astFile.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.CONST {
			for _, spec := range d.Specs {
				if vSpec, ok := spec.(*ast.ValueSpec); ok {
					if vSpecType, ok := vSpec.Type.(*ast.Ident); ok && vSpecType.Name == t {
						if vSpecValue, ok := vSpec.Values[0].(*ast.BasicLit); ok {
							enum.EnumFields = append(enum.EnumFields, &EnumField{
								Name:  vSpec.Names[0].Name,
								Value: vSpecValue.Value,
							})
						}
					}
				}
			}
		}
	}

	if len(enum.EnumFields) == 0 {
		return nil, errors.New("must specify enum fields")
	}

	return enum, nil
}

func getSearchPath(pkgName, repoImport, repoPath string, astFile *ast.File) (string, error) {
	importValue, err := utils.GetImportPath(pkgName, astFile)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(repoImport, importValue)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoPath, rel), nil
}
