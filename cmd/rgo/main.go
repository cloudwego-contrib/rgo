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
	"log"
	"os"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/urfave/cli/v2"
)

func main() {
	client := Init()

	err := client.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func Init() *cli.App {
	verboseFlag := cli.BoolFlag{Name: "verbose,vv", Usage: "turn on verbose mode"}

	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = Name
	app.Usage = AppUsage
	app.Version = Version
	// The default separator for multiple parameters is modified to ";"
	app.SliceFlagSeparator = ";"

	// global flags
	app.Flags = []cli.Flag{
		&verboseFlag,
	}

	// Commands
	app.Commands = []*cli.Command{
		{
			Name:  GenerateName,
			Usage: GenerateUsage,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "config,c", Usage: "rgo_config file path", Destination: &idlConfigPath, DefaultText: consts.RGOConfigPath},
			},
			Action: func(c *cli.Context) error {
				return GenerateRGOCode()
			},
		},
		{
			Name:  CleanName,
			Usage: CleanUsage,
			Flags: nil,
			Action: func(c *cli.Context) error {
				return Clean()
			},
		},
		{
			Name:  InitName,
			Usage: InitUsage,
			Flags: []cli.Flag{
				&cli.StringFlag{Name: consts.TypeFlag, Usage: "ide type, default: vscode", DefaultText: consts.VSCode},
			},
			Action: InitProject,
		},
	}
	return app
}

const (
	AppUsage = "generate or clean rpc code for rgo"

	GenerateName  = "generate"
	GenerateUsage = `generate RPC code for rgo

Examples:
  # Generate rgo code 
  rgo generate
`

	CleanName  = "clean"
	CleanUsage = `clean rgo code

Examples:
  # Clean rgo code 
  rgo clean
`
	InitName  = "init_config"
	InitUsage = `init rgo project config
Examples:
	# Init rgo project
	rgo init_config
`
)
