package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/sdk"
	"os"
	"os/exec"
	"path/filepath"
)

func generateThriftCode(idlPath, rgoSrcPath string) error {
	outputDir := rgoSrcPath
	command := "thriftgo"
	args := []string{
		"-g", "go:template=slim,gen_deep_equal=false,gen_setter=false,no_default_serdes" + fmt.Sprintf(",package_prefix=%s", filepath.Join(consts.RGOModuleName, extractPathAfterCache(rgoSrcPath))),
		"-o", outputDir,
		"--recurse",
		idlPath,
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Executing command: %s %v\n", command, args)
	err := cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("thriftgo execution failed: %v", err))
	}

	return nil
}

func (rg *RGOGenerator) GenRgoBaseCode(idlPath, rgoSrcPath string) error {
	outputDir := filepath.Join(rgoSrcPath, "kitex_gen")

	args := []string{
		"-g", "go:template=slim,gen_deep_equal=false,gen_setter=false,no_default_serdes,no_fmt" + fmt.Sprintf(",package_prefix=%s", filepath.Join(consts.RGOModuleName, "kitex_gen")),
		"-o", outputDir,
		"--recurse",
		idlPath,
	}

	path, _ := filepath.Abs("./")

	err := sdk.RunThriftgoAsSDK(path, nil, args...)
	if err != nil {
		return errors.New(fmt.Sprintf("thriftgo execution failed: %v", err))
	}

	return nil
}

func parseIDLFile(idlFile string) (*parser.Thrift, error) {
	thriftFile, err := parser.ParseFile(idlFile, nil, true)
	if err != nil {
		return nil, err
	}

	return thriftFile, nil
}
