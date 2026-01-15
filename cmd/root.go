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

// Package cmd defines the CLI entrypoints and command configuration.
package cmd

import (
	"os"

	"github.com/okdp/spark-web-proxy/internal/config"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd is the root CLI command for spark-web-proxy.
var RootCmd = &cobra.Command{
	Use:   "spark-web-proxy",
	Short: "Spark UI Proxy",
	Run:   runSparkUIController,
}

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("proxy.listenAddress", "localhost")
	viper.SetDefault("proxy.port", 8090)
	viper.SetDefault("proxy.mode", "release")

	viper.SetDefault("spark.history.scheme", "http")
	viper.SetDefault("spark.history.service", "localhost")
	viper.SetDefault("spark.history.port", 18080)

	viper.SetDefault("spark.ui.proxyBase", "/sparkui")
	viper.SetDefault("spark.jobNamespaces", "default")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "console")

	viper.SetDefault("security.cors.allowedOrigins", []string{"*"})
	viper.SetDefault("security.cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"})
	viper.SetDefault("security.cors.allowedHeaders", []string{"*"})
	viper.SetDefault("security.cors.exposedHeaders", []string{"Content-Length"})
	viper.SetDefault("security.cors.allowCredentials", false)
	viper.SetDefault("security.cors.maxAge", 3600)

	viper.SetConfigName("spark-web-proxy")
	viper.SetConfigType("yaml")

	RootCmd.PersistentFlags().String("config", "config.yaml", "Path to configuration file")
	if err := viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config")); err != nil {
		panic("Unable to read server configuration: " + err.Error())
	}
}

// Execute runs the root command and exits with a non-zero status on failure.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

// runSparkUIController starts the Spark Web Proxy server using the loaded configuration.
func runSparkUIController(_ *cobra.Command, _ []string) {
	config := config.GetAppConfig()
	log.SetupGlobalLogger(config.Logging)

	server := server.NewSparkUIProxyServer(config)
	log.Info("ListenAddress %s: ", config.Proxy.ListenAddress)
	log.Info("Port %d: ", config.Proxy.Port)
	log.Info("spark ui proxy started on port %d", config.Proxy.Port)
	log.Fatal(server.ListenAndServe())
}
