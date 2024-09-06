package config

type GoWork struct {
	Go  string     `json:"Go"`
	Use []UseEntry `json:"Use"`
}

type UseEntry struct {
	DiskPath string `json:"DiskPath"`
}
