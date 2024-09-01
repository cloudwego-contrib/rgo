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

package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bytedance/sonic"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"golang.org/x/tools/go/packages"
)

func (rg *RGOGenerator) generateRGOPackages(formatServiceName, path string) error {
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

	outputFile := filepath.Join(rg.RGOBasePath, consts.PkgMetaPath, formatServiceName, "rgo_packages.json")

	err = os.MkdirAll(filepath.Dir(outputFile), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	err = os.WriteFile(outputFile, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	return nil
}
