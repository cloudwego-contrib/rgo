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
	"github.com/cloudwego-contrib/rgo/pkg/rlog"
	"go.uber.org/zap"
	"os/exec"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/cloudwego/thriftgo/parser"
)

func (rg *RGOGenerator) GenRgoBaseCode(module, serviceName, formatServiceName, idlPath, rgoSrcPath string) error {
	customArgs := []string{
		"-frugal-pretouch",
		"-thrift", "template=slim",
		"-thrift", "frugal_tag",
		"-thrift", "gen_deep_equal=false",
		"-thrift", "gen_setter=false",
		"-thrift", "no_default_serdes",
		"-thrift", "no_fmt",
	}

	args := []string{
		"kitex",
		fmt.Sprintf("--%s", consts.PluginTypeFlag), consts.EditPeriod,
		fmt.Sprintf("--%s", consts.PwdFlag), rgoSrcPath,
		fmt.Sprintf("--%s", consts.ModuleFlag), module,
		fmt.Sprintf("--%s", consts.ServiceNameFlag), serviceName,
		fmt.Sprintf("--%s", consts.FormatServiceNameFlag), formatServiceName,
		fmt.Sprintf("--%s", consts.IDLPathFlag), idlPath,
	}

	for _, customArg := range customArgs {
		args = append(args, fmt.Sprintf("--%s", consts.KitexArgsFlag), customArg)
	}

	rlog.Debug("generate rgo base code", zap.Any("args", args))

	cmd := exec.Command("rgo", args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error generate rgo base code: %v", err)
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
