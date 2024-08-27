package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego-contrib/rgo/driver/internal/utils"
	"golang.org/x/tools/go/packages"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

const (
	RGOBasePath = "RGO_BASE_PATH"
)

type (
	DefaultPackageLoader struct {
		Ret *packages.DriverResponse
	}

	LoadMode int

	DriverRequest struct {
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
)

var (
	rgoBasePath string

	_ GoPackagesDriverLoad = (*DefaultPackageLoader)(nil)
)

type GoPackagesDriverLoad interface {
	// pkgs is a list of package patterns to load.
	LoadPackages(cfg *packages.Config, pkgs ...string) []*packages.Package
}

func (t *DefaultPackageLoader) LoadPackages(cfg *packages.Config, pkgs ...string) []*packages.Package {
	for _, pkg := range pkgs {
		p, _ := packages.Load(cfg, pkg)
		t.Ret.Packages = append(t.Ret.Packages, p...)
	}
	return t.Ret.Packages
}

func init() {
	rgoBasePath = os.Getenv(RGOBasePath)
	if rgoBasePath == "" {
		rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), ".RGO", "cache")
	}

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
	if err := json.NewDecoder(in).Decode(&req); err != nil {
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
		return fmt.Errorf("failed to get default driver response: %v", err)
	}

	for k := len(ret.Packages) - 1; k >= 0; k-- {
		if len(ret.Packages[k].Errors) > 0 && strings.HasPrefix(ret.Packages[k].PkgPath, "rgo/") {
			ret.Packages = append(ret.Packages[:k], ret.Packages[k+1:]...)
		}
	}

	basePath := filepath.Join(rgoBasePath, "pkg_meta")

	curWorkPath, err := utils.GetCurrentPathWithUnderline()

	targetPath := filepath.Join(basePath, curWorkPath)

	targetPkgs, err = getTargetPackages(targetPath)
	if err != nil {
		log.Printf("Error getting target packages from path %s: %v", targetPath, err)
	}

	for _, pkg := range targetPkgs {
		ret.Roots = append(ret.Roots, pkg.ID)
		ret.Packages = append(ret.Packages, pkg)
	}

	var loader DefaultPackageLoader
	loader.Ret = ret
	loader.LoadPackages(cfg, "context", "fmt", "github.com/cloudwego/kitex/client", "github.com/cloudwego/kitex/client/callopt")

	data, err := sonic.Marshal(ret)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err.Error())
	}

	_, err = out.Write(data)
	return err
}

func getTargetPackages(path string) ([]*packages.Package, error) {
	var results []*packages.Package

	// Check if the path directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return results, nil
	}

	// Read all subdirectories and files under the path directory
	directories, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	// Traverse the first level subdirectory
	for _, dir := range directories {
		if dir.IsDir() {
			jsonFilePath := filepath.Join(path, dir.Name(), "rgo_packages.json")

			if _, err := os.Stat(jsonFilePath); err == nil {
				data, err := os.ReadFile(jsonFilePath)
				if err != nil {
					return nil, fmt.Errorf("failed to read json file %s: %v", jsonFilePath, err)
				}

				var response []*packages.Package
				if err := sonic.Unmarshal(data, &response); err != nil {
					return nil, fmt.Errorf("failed to parse json file %s: %v", jsonFilePath, err)
				}

				results = append(results, response...)
			}
		}
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
