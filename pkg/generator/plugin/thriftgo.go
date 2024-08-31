package plugin

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego-contrib/rgo/pkg/utils"
	"github.com/cloudwego/thriftgo/plugin"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
)

func GetRGOThriftgoPlugin(pwd, formatServiceName string, Args []string) (*RGOThriftgoPlugin, error) {
	rgoPlugin := &RGOThriftgoPlugin{}

	rgoPlugin.Pwd = pwd
	rgoPlugin.FormatServiceName = formatServiceName
	rgoPlugin.Args = Args

	return rgoPlugin, nil
}

type RGOThriftgoPlugin struct {
	Args              []string
	FormatServiceName string
	Pwd               string
}

func (r *RGOThriftgoPlugin) GetName() string {
	return "rgo"
}

func (r *RGOThriftgoPlugin) GetPluginParameters() []string {
	return r.Args
}

func (r *RGOThriftgoPlugin) Invoke(req *plugin.Request) (res *plugin.Response) {
	formatServiceName := r.FormatServiceName

	thrift := req.AST

	fset := token.NewFileSet()

	// Call the function
	file, err := BuildRGOThriftAstFile(formatServiceName, thrift)

	exist, err := utils.FileExistsInPath(r.Pwd, "go.mod")
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(err.Error()),
		}
	}

	if !exist {
		err = os.MkdirAll(r.Pwd, os.ModePerm)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(fmt.Sprintf("failed to create directory: %v", err)),
			}
		}

		err = utils.InitGoMod(filepath.Join(consts.RGOModuleName, formatServiceName), r.Pwd)
		if err != nil {
			return &plugin.Response{
				Error: strToPointer(err.Error()),
			}
		}
	}

	outputFile, err := os.Create(filepath.Join(r.Pwd, "rgo_cli.go"))
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to create file: %v", err)),
		}
	}
	defer outputFile.Close()

	if err = format.Node(outputFile, fset, file); err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to format file: %v", err)),
		}
	}

	err = utils.RunGoModTidyInDir(r.Pwd)
	if err != nil {
		return &plugin.Response{
			Error: strToPointer(fmt.Sprintf("failed to go mod tidy: %v", err)),
		}
	}

	return &plugin.Response{}
}
