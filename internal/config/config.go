/*
 *    Copyright 2025 The OKDP Authors.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

// Package config provides application configuration loading and access
// using Viper, including support for dynamic reloads.
package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/okdp/spark-web-proxy/internal/utils"
)

// ApplicationConfig represents the root configuration of the application.
type ApplicationConfig struct {
	Proxy    Proxy    `mapstructure:"proxy"`
	Spark    Spark    `mapstructure:"spark"`
	Security Security `mapstructure:"security"`
	Logging  Logging  `mapstructure:"logging"`
}

// Proxy defines the reverse proxy server configuration.
type Proxy struct {
	ListenAddress string `mapstructure:"listenAddress"`
	Port          int    `mapstructure:"port"`
	Mode          string `mapstructure:"mode"`
}

// Spark defines Spark-related configuration.
type Spark struct {
	History       History  `mapstructure:"history"`
	UI            UI       `mapstructure:"ui"`
	JobNamespaces []string `json:"jobNamespaces"`
}

// History defines Spark History Server configuration.
type History struct {
	Scheme  string `yaml:"scheme"`
	Service string `yaml:"service"`
	Port    int    `yaml:"port"`
}

// UI defines Spark UI configuration
type UI struct {
	ProxyBase string `yaml:"proxyBase"`
}

// Logging configuration
type Logging struct {
	Level  string `yaml:"provider"`
	Format string `yaml:"format"`
}

// Security configuration
type Security struct {
	Cors    Cors              `yaml:"cors"`
	Headers map[string]string `yaml:"headers"`
}

// Cors configuration
type Cors struct {
	AllowedOrigins   []string `json:"allowedOrigins"`
	AllowedMethods   []string `json:"allowedMethods"`
	AllowedHeaders   []string `json:"allowedHeaders"`
	ExposedHeaders   []string `json:"exposedHeaders"`
	AllowCredentials bool     `json:"allowCredentials"`
	MaxAge           int64    `json:"maxAge"`
}

var (
	instance *ApplicationConfig
	once     sync.Once
)

// GetAppConfig returns a singleton instance of the application configuration.
// It reads the yaml file provided in the argument (--config=/path/to/app-config.yaml) at the startup of the application
// into the ApplicationConfig struct
func GetAppConfig() *ApplicationConfig {
	once.Do(func() {
		instance = &ApplicationConfig{}
		configFile := viper.GetString("config")
		viper.SetConfigFile(configFile)
		fmt.Println("Loading configuration from config file: ", configFile)

		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("failed to read the configuration file")
			panic(err)
		}

		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("Config file changed:", e.Name)
			if err := viper.Unmarshal(&instance); err != nil {
				fmt.Println("failed to register config change watcher")
				panic(err)
			}
		})

		if err := viper.Unmarshal(&instance); err != nil {
			fmt.Println("failed to parse the configuration file")
			panic(err)
		}

		printConfig(configFile)
	})
	return instance
}

// GetSparkHistoryBaseURL returns the base URL of the Spark History Server
// constructed from the configured scheme, service, and port.
// It validates the resulting URL and panics if it is invalid.
func (c ApplicationConfig) GetSparkHistoryBaseURL() string {
	sparkHistoryBaseURL := fmt.Sprintf("%s://%s:%d", c.Spark.History.Scheme,
		c.Spark.History.Service,
		c.Spark.History.Port)

	utils.ValidateURL(sparkHistoryBaseURL, fmt.Sprintf("The Spark History Server URL is not valid (Scheme: %s, Service: %s, Port: %d)",
		c.Spark.History.Scheme,
		c.Spark.History.Service,
		c.Spark.History.Port))

	return sparkHistoryBaseURL
}

func printConfig(fileConfigPath string) {
	content, err := os.ReadFile(fileConfigPath)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
	fmt.Println()
}
