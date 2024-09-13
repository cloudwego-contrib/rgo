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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/rlog"

	"github.com/cloudwego-contrib/rgo/pkg/utils"

	"github.com/cloudwego-contrib/rgo/cmd/rgopackagesdriver/internal"
	"golang.org/x/tools/go/packages"
)

const (
	RGOBasePath = ".rgo/cache"
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
	// LoadPackages pkgs is a list of package patterns to load.
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
	curWorkPath, err := utils.GetProjectHashPathWithUnderline()
	if err != nil {
		panic(err)
	}

	rgoBasePath = filepath.Join(utils.GetDefaultUserPath(), RGOBasePath, curWorkPath)
}

func main() {
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()

	rlog.InitLogger(filepath.Join(rgoBasePath, consts.LogPath, consts.RGOPackagesDriver))

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
	if err = json.NewDecoder(in).Decode(&req); err != nil {
		rlog.Errorf("unable to decode driver request: %v", err)
		return err
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

	ret, b, err := internal.UnsafeGetDefaultDriverResponse(cfg, args...)
	if err != nil || b {
		rlog.Errorf("failed to get default driver response: %v", err)
		return err
	}

	for k := len(ret.Packages) - 1; k >= 0; k-- {
		if len(ret.Packages[k].Errors) > 0 && strings.HasPrefix(ret.Packages[k].PkgPath, "rgo/") {
			ret.Packages = append(ret.Packages[:k], ret.Packages[k+1:]...)
		}
	}

	targetPath := filepath.Join(rgoBasePath, "pkg_meta")

	targetPkgs, err = getTargetPackages(targetPath)
	if err != nil {
		rlog.Warnf("unable get target packages from path %s: %v", targetPath, err)
	}

	for _, pkg := range targetPkgs {
		ret.Roots = append(ret.Roots, pkg.ID)
	}

	ret.Packages = append(targetPkgs, ret.Packages...)

	var loader DefaultPackageLoader
	loader.Ret = ret
	loader.LoadPackages(cfg, "context", "fmt", "github.com/cloudwego/kitex/client", "github.com/cloudwego/kitex/client/callopt")

	data, err := json.Marshal(ret)
	if err != nil {
		rlog.Errorf("json marshal error: %v", err.Error())
		return err
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
				if err := json.Unmarshal(data, &response); err != nil {
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
