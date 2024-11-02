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

type IDLRepo struct {
	RepoName string `yaml:"repo_name" mapstructure:"repo_name"`
	GitUrl   string `yaml:"git_url" mapstructure:"git_url"`
	Branch   string `yaml:"branch" mapstructure:"branch"`
	Commit   string `yaml:"commit" mapstructure:"commit"`
}

type IDL struct {
	ServiceName       string `yaml:"service_name" mapstructure:"service_name"`
	FormatServiceName string
	IDLPath           string `yaml:"idl_path" mapstructure:"idl_path"`
	RepoName          string `yaml:"repo_name" mapstructure:"repo_name"`
}

type RGOConfig struct {
	Auth          AuthConfig `yaml:"auth" mapstructure:"auth"`
	Mode          string     `yaml:"mode" mapstructure:"mode"`
	ProjectModule string     `yaml:"project_module" mapstructure:"project_module"`
	IDLRepos      []IDLRepo  `yaml:"idl_repos" mapstructure:"idl_repos"`
	IDLs          []IDL      `yaml:"idls" mapstructure:"idls"`
}
