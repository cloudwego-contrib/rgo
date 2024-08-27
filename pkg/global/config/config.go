package config

type IDLRepo struct {
	RepoName string `yaml:"repo_name" mapstructure:"repo_name"`
	RepoGit  string `yaml:"repo_git" mapstructure:"repo_git"`
	Branch   string `yaml:"branch" mapstructure:"branch"`
	Commit   string `yaml:"commit" mapstructure:"commit"`
}

type IDL struct {
	ServiceAddress string `yaml:"service_address" mapstructure:"service_address"`
	ServiceName    string `yaml:"service_name" mapstructure:"service_name"`
	IDLPath        string `yaml:"idl_path" mapstructure:"idl_path"`
	IDLRepo        string `yaml:"idl_repo" mapstructure:"idl_repo"`
}

type RGOConfig struct {
	IDLRepos []IDLRepo `yaml:"idl_repos" mapstructure:"idl_repos"`
	IDLs     []IDL     `yaml:"idls" mapstructure:"idls"`
}
