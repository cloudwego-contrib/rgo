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

import (
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func ReadConfig(path string) (*RGOConfig, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file %s: %v", path, err)
	}

	// Read Config
	c := &RGOConfig{}
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatalf("Failed to parse config into struct: %v", err)
	}

	for i := range c.IDLs {
		c.IDLs[i].FormatServiceName = strings.ReplaceAll(c.IDLs[i].ServiceName, "-", "_")
		c.IDLs[i].FormatServiceName = strings.ReplaceAll(c.IDLs[i].FormatServiceName, ".", "_")
	}

	return c, nil
}

var ConfigChangeHandler func(e fsnotify.Event)

func RewriteRGOConfig(key string, value interface{}) error {
	viper.OnConfigChange(nil)

	viper.Set(key, value)

	err := viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	viper.OnConfigChange(ConfigChangeHandler)

	return nil
}
