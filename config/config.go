package config

type IDLRepo struct {
	Repo   string `yaml:"repo" mapstructure:"repo"`
	Branch string `yaml:"branch" mapstructure:"branch"`
	Commit string `yaml:"commit" mapstructure:"commit"`
}

type Config struct {
	ServiceName    string `yaml:"service_name" mapstructure:"service_name"`
	ServiceAddress string `yaml:"service_address" mapstructure:"service_address"`
}

type IDL struct {
	IDLRepo string   `yaml:"idl_repo" mapstructure:"idl_repo"`
	IDLPath string   `yaml:"idl_path" mapstructure:"idl_path"`
	Config  []Config `yaml:"config" mapstructure:"config"`
}

type RGOConfig struct {
	IDLRepos map[string]IDLRepo `yaml:"idl_repos" mapstructure:"idl_repos"`
	IDLs     []IDL              `yaml:"idls" mapstructure:"idls"`
}
