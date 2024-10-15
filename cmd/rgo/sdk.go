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
	pluginType := c.String(consts.PluginTypeFlag)
	thriftgoCustomArgs := c.StringSlice(consts.ThriftgoCustomArgsFlag)

	if pluginType == "" {
		err := sdk.RunThriftgoAsSDK(pwd, nil, thriftgoCustomArgs...)
		if err != nil {
			return err
		}
	} else {
		rgoPlugin, err := plugin.GetRGOPlugin(pluginType, pwd, module, serviceName, formatServiceName)
		if err != nil {
			return err
		}
		err = sdk.RunThriftgoAsSDK(pwd, []plugin2.SDKPlugin{rgoPlugin}, thriftgoCustomArgs...)
		if err != nil {
			return err
		}

	}

	return nil
}

func RunKitexCommand(c *cli.Context) error {
	pwd := c.String(consts.PwdFlag)
	module := c.String(consts.ModuleFlag)
	serviceName := c.String(consts.ServiceNameFlag)
	formatServiceName := c.String(consts.FormatServiceNameFlag)
	idlPath := c.String(consts.IDLPathFlag)
	pluginType := c.String(consts.PluginTypeFlag)
	kitexCustomArgs := c.StringSlice(consts.KitexArgsFlag)

	var rgoPlugin plugin2.SDKPlugin
	var err error

	rgoPlugin, err = plugin.GetRGOPlugin(pluginType, pwd, module, serviceName, formatServiceName)
	if err != nil {
		return err
	}

	err = generateKitexGen(pwd, module, idlPath, kitexCustomArgs, rgoPlugin)
	if err != nil {
		return fmt.Errorf("failed to generate rgo code:%v", err)
	}

	return nil
}
