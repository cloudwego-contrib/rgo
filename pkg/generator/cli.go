package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/generator/plugin"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
)

func (rg *RGOGenerator) GenerateRGOCode(formatServiceName, idlPath, rgoSrcPath string) error {
	exist, err := utils.FileExistsInPath(rgoSrcPath, "go.mod")
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(rgoSrcPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		err = utils.InitGoMod(filepath.Join(consts.RGOModuleName, formatServiceName), rgoSrcPath)
		if err != nil {
			return err
		}
	}

	fileType := filepath.Ext(idlPath)

	switch fileType {
	case ".thrift":
		err = rg.GenRgoBaseCode(formatServiceName, idlPath, rgoSrcPath)
		if err != nil {
			return err
		}

		return rg.generateRGOPackages(formatServiceName, rgoSrcPath)
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

	f, err := plugin.BuildRGOThriftAstFile(serviceName, thriftFile)
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
