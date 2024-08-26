package main

import (
	"context"
	"fmt"
	"github.com/cloudwego-contrib/rgo/driver/internal/utils"
	"golang.org/x/tools/go/packages"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

type LoadMode int

type DriverRequest struct {
	Mode LoadMode `json:"mode"`

	// Env specifies the environment the underlying build system should be run in.
	Env []string `json:"env"`

	// BuildFlags are flags that should be passed to the underlying build system.
	BuildFlags []string `json:"build_flags"`

	// Tests specifies whether the patterns should also return test packages.
	Tests bool `json:"tests"`

	// Overlay maps file paths (relative to the driver's working directory)
	// to the contents of overlay files (see Config.Overlay).
	Overlay map[string][]byte `json:"overlay"`
}

var rgoRepoPath string

// todo: 从配置文件中读取
func init() {
	rgoRepoPath = "/Users/violapioggia/RGO"
}

func main() {
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Stdin, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		// gopls will check the packages driver exit code, and if there is an
		// error, it will fall back to go list. Obviously we don't want that,
		// so force a 0 exit code.
		os.Exit(0)
	}
}

func run(ctx context.Context, in io.Reader, out io.Writer, args []string) error {
	var (
		targetPkgs []*packages.Package
		err        error
	)

	req := &DriverRequest{}
	if err := sonic.NewDecoder(in).Decode(&req); err != nil {
		return fmt.Errorf("unable to decode driver request: %w", err)
	}

	for k := len(req.Env) - 1; k >= 0; k-- {
		if strings.Contains(req.Env[k], "GOPACKAGESDRIVER") {
			req.Env = append(req.Env[:k], req.Env[k+1:]...)
			break
		}
	}

	cfg := &packages.Config{
		Mode:       packages.LoadMode(req.Mode),
		Context:    ctx,
		Env:        req.Env,
		Overlay:    req.Overlay,
		Tests:      req.Tests,
		BuildFlags: req.BuildFlags,
	}

	ret, b, err := utils.UnsafeGetDefaultDriverResponse(cfg, args...)
	if err != nil || b {
		panic(err)
	}

	for k := len(ret.Packages) - 1; k >= 0; k-- {
		if len(ret.Packages[k].Errors) > 0 && strings.HasPrefix(ret.Packages[k].PkgPath, "rgo/") {
			ret.Packages = append(ret.Packages[:k], ret.Packages[k+1:]...)
		}
	}

	//todo: 找当前目录
	basePath := filepath.Join(rgoRepoPath, "cache", "pkg_meta")

	targetPkgs, err = getTargetPackages(basePath)
	if err != nil {
		//TODO: log error
	}

	for _, pkg := range targetPkgs {
		ret.Roots = append(ret.Roots, pkg.ID)
		ret.Packages = append(ret.Packages, pkg)
	}

	//todo: 添加注入依赖包
	ctxPackage, _ := packages.Load(cfg, "context")
	ret.Packages = append(ret.Packages, ctxPackage...)

	fmtPackage, _ := packages.Load(cfg, "fmt")
	ret.Packages = append(ret.Packages, fmtPackage...)

	kitexClientPackage, _ := packages.Load(cfg, "github.com/cloudwego/kitex/client")
	ret.Packages = append(ret.Packages, kitexClientPackage...)

	kitexClientOptPackage, _ := packages.Load(cfg, "github.com/cloudwego/kitex/client/callopt")
	ret.Packages = append(ret.Packages, kitexClientOptPackage...)

	data, err := sonic.Marshal(ret)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err.Error())
	}

	_, err = out.Write(data)
	return err
}

func getTargetPackages(basePath string) ([]*packages.Package, error) {
	var results []*packages.Package

	// 递归函数，用于查找目标文件
	var findPackages func(string) error
	findPackages = func(path string) error {
		directories, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("failed to read directory: %v", err)
		}

		for _, dir := range directories {
			if dir.IsDir() {
				// 递归处理子目录
				if err := findPackages(filepath.Join(path, dir.Name())); err != nil {
					return err
				}
			} else if dir.Name() == "rgo_packages.json" {
				// 找到目标文件，处理它
				jsonFilePath := filepath.Join(path, dir.Name())
				data, err := os.ReadFile(jsonFilePath)
				if err != nil {
					return fmt.Errorf("failed to read json file %s: %v", jsonFilePath, err)
				}

				var response []*packages.Package
				if err := sonic.Unmarshal(data, &response); err != nil {
					return fmt.Errorf("failed to parse json file %s: %v", jsonFilePath, err)
				}

				results = append(results, response...)
			}
		}

		return nil
	}

	// 从基目录开始递归查找
	if err := findPackages(basePath); err != nil {
		return nil, err
	}

	return results, nil
}

func signalContext(parentCtx context.Context, signals ...os.Signal) (ctx context.Context, stop context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)
	ch := make(chan os.Signal, 1)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	signal.Notify(ch, signals...)

	return ctx, cancel
}
