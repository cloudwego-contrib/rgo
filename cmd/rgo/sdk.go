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
	"fmt"

	"github.com/cloudwego-contrib/rgo/pkg/consts"
	"github.com/cloudwego-contrib/rgo/pkg/generator/plugin"
	plugin2 "github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/sdk"
	"github.com/urfave/cli/v2"
)

func RunThriftgoCommand(c *cli.Context) error {
	pwd := c.String(consts.PwdFlag)
	module := c.String(consts.ModuleFlag)
	serviceName := c.String(consts.ServiceNameFlag)
	formatServiceName := c.String(consts.FormatServiceNameFlag)
	thriftgoCustomArgs := c.StringSlice(consts.ThriftgoCustomArgsFlag)

	rgoPlugin, err := plugin.GetRGOThriftgoPlugin(pwd, module, serviceName, formatServiceName, nil)
	if err != nil {
		return err
	}

	err = sdk.RunThriftgoAsSDK(pwd, []plugin2.SDKPlugin{rgoPlugin}, thriftgoCustomArgs...)
	if err != nil {
		return fmt.Errorf("thriftgo execution failed: %v", err)
	}

	return nil
}
