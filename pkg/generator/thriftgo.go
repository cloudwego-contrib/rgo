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
	"path/filepath"

	plugin2 "github.com/cloudwego-contrib/rgo/pkg/generator/plugin"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/sdk"
)

func (rg *RGOGenerator) GenRgoBaseCode(formatServiceName, idlPath, rgoSrcPath string) error {
	outputDir := filepath.Join(rgoSrcPath, "kitex_gen")

	args := []string{
		"-g", "go:template=slim,gen_deep_equal=false,gen_setter=false,no_default_serdes,no_fmt" + fmt.Sprintf(",package_prefix=%s", filepath.Join(consts.RGOModuleName, "kitex_gen")),
		"-o", outputDir,
		"--recurse",
		idlPath,
	}

	rgoPlugin, err := plugin2.GetRGOThriftgoPlugin(rgoSrcPath, formatServiceName, nil)
	if err != nil {
		return err
	}

	err = sdk.RunThriftgoAsSDK(rgoSrcPath, []plugin.SDKPlugin{rgoPlugin}, args...)
	if err != nil {
		return fmt.Errorf("thriftgo execution failed: %v", err)
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
