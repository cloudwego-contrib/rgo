package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/rgo/consts"
	"github.com/cloudwego/thriftgo/parser"
	"os"
	"os/exec"
	"path/filepath"
)

func generateThriftCode(idlRepoPath, idlPath, rgoRepoPath string) error {
	outputDir := filepath.Join(rgoRepoPath, "src/rgo-gen-go", idlRepoPath, "{namespace}")
	command := "thriftgo"
	args := []string{
		"-g", "go:template=slim,gen_deep_equal=false,gen_setter=false,no_default_serdes" + fmt.Sprintf(",package_prefix=%s", filepath.Join(consts.RGOModuleName, consts.RGOGenCodePath, idlRepoPath)),
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

func parseIDLFile(idlFile string) (*parser.Thrift, error) {
	thriftFile, err := parser.ParseFile(idlFile, nil, true)
	if err != nil {
		return nil, err
	}

	return thriftFile, nil
}
