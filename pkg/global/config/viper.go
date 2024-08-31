package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"strings"
)

func ReadConfig(path string) (*RGOConfig, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file %s: %v", path, err)
	}

	// 解析配置文件内容到结构体
	c := &RGOConfig{}
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatalf("Failed to parse config into struct: %v", err)
	}

	for i := range c.IDLs {
		strings.ReplaceAll(c.IDLs[i].ServiceName, "-", "_")
		strings.ReplaceAll(c.IDLs[i].ServiceName, ".", "_")
	}

	return c, nil
}

var ConfigChangeHandler func(e fsnotify.Event)

func RewriteRGOConfig(key string, value interface{}) error {
	viper.OnConfigChange(nil)

	viper.Set(key, value)

	err := viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	viper.OnConfigChange(ConfigChangeHandler)

	return nil
}
