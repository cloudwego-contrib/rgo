package generator

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
)

type RGOPackages struct {
	Packages []packages.Package `json:"packages"`
}

func (rg *RGOGenerator) generateRGOPackages(curWorkPath, serviceName, path string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypesSizes |
			packages.NeedModule |
			packages.NeedEmbedFiles,
		Dir: path,
	}

	pkgs, err := packages.Load(cfg, filepath.Join(path, "..."))
	if err != nil {
		return fmt.Errorf("failed to load packages: %v", err)
	}

	Packages := make([]*packages.Package, 0)

	Packages = append(Packages, pkgs...)

	data, err := sonic.Marshal(Packages)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	outputFile := filepath.Join(rg.RGOBasePath, consts.PkgMetaPath, curWorkPath, serviceName, "rgo_packages.json")

	err = os.MkdirAll(filepath.Dir(outputFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	return nil

}
