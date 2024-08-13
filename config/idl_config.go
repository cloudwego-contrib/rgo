package config

type RgoInfo struct {
	IDLConfig []IDLConfig `yaml:"idl"`
}

type IDLConfig struct {
	Repository     string `yaml:"repository"`
	IDLPath        string `yaml:"idl_path"`
	ServiceName    string `yaml:"service_name"`
	ServiceAddress string `yaml:"service_address"`
	Branch         string `yaml:"branch,omitempty"` // Branch 是可选字段
}
