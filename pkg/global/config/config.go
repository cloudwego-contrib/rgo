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
	IDLRepos []IDLRepo `yaml:"idl_repos" mapstructure:"idl_repos"`
	IDLs     []IDL     `yaml:"idls" mapstructure:"idls"`
}
