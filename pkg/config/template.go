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

package config

import "github.com/cloudwego/thriftgo/parser"

type RGOClientTemplateData struct {
	RGOModuleName     string   // Name of the RGO module (e.g., rgo)
	ServiceName       string   // Name of the service (e.g., service.one)
	FormatServiceName string   // Formatted service name (e.g., service_one)
	Imports           []string // List of imports required for the client (e.g., context, github.com/cloudwego/kitex/client)
	*parser.Thrift
}
