package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
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

	go watchConfig(c)

	return c, nil
}

func watchConfig(c *RGOConfig) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		// 重新解析配置文件内容到结构体
		if err := viper.Unmarshal(&c); err != nil {
			log.Printf("Failed to parse updated config into struct: %v", err)
		} else {
			// 打印更新后的配置
			fmt.Printf("Updated Config: %+v\n", c)
		}
	})

	// 保持程序运行状态，持续监听配置文件变化
	select {}
}
