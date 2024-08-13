package generator

import (
	"errors"
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"os"
	"os/exec"
	"path/filepath"
)

func generateThriftCode(idlPath, repoPath string) error {
	outputDir := filepath.Join(repoPath, "src/rgo-gen-go/{namespace}")
	command := "thriftgo"
	args := []string{
		"-g", "go:template=slim,gen_deep_equal=false,gen_setter=false,no_default_serdes",
		"-o", outputDir,
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
	content, err := os.ReadFile(idlFile)
	if err != nil {
		return nil, err
	}

	thriftFile, err := parser.ParseString(idlFile, string(content))
	if err != nil {
		return nil, err
	}

	return thriftFile, nil
}
