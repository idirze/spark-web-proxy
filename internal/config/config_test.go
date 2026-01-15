/*
 *    Copyright 2024 The OKDP Authors.
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

package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig_Server(t *testing.T) {
	// Given
	viper.Set("config", "testdata/application.yaml")
	// When
	proxy := GetAppConfig().Proxy
	// Then
	assert.Equal(t, "0.0.0.0", proxy.ListenAddress, "ListenAddress")
	assert.Equal(t, 8090, proxy.Port, "Port")
	assert.Equal(t, "debug", proxy.Mode, "Mode")
}

func Test_LoadConfig_Server_Logging(t *testing.T) {
	// Given
	viper.Set("config", "testdata/application.yaml")
	// When
	logging := GetAppConfig().Logging
	// Then
	assert.Equal(t, "debug", logging.Level, "Level")
	assert.Equal(t, "console", logging.Format, "Format")
}

func Test_LoadConfig_Server_Cors(t *testing.T) {
	// Given
	viper.Set("config", "testdata/application.yaml")
	// When
	cors := GetAppConfig().Security.Cors
	// Then
	assert.Equal(t, []string{"*"}, cors.AllowedOrigins, "AllowedOrigins")
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "PATCH"}, cors.AllowedMethods, "AllowedMethods")
	assert.Equal(t, []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}, cors.AllowedHeaders, "AllowedHeaders")
	assert.Equal(t, []string{"Content-Length"}, cors.ExposedHeaders, "ExposedHeaders")
	assert.True(t, true, cors.AllowCredentials, "AllowCredentials")
	assert.Equal(t, int64(3600), cors.MaxAge, "MaxAge")
}

func Test_LoadConfig_Server_Headers(t *testing.T) {
	// Given
	viper.Set("config", "testdata/application.yaml")
	// When
	security := GetAppConfig().Security
	// Then
	assert.Equal(t, map[string]string{"x-frame-options": "DENY", "x-content-type-options": "nosniff"}, security.Headers, "Headers")
}

func Test_LoadConfig_Spark(t *testing.T) {
	// Given
	viper.Set("config", "testdata/application.yaml")
	// When
	spark := GetAppConfig().Spark
	// Then
	assert.Equal(t, "http", spark.History.Scheme, "spark.history.scheme")
	assert.Equal(t, "spark-history-server", spark.History.Service, "spark.history.service")
	assert.Equal(t, 18080, spark.History.Port, "spark.history.port")

	assert.Equal(t, "/sparkui", spark.UI.ProxyBase, "spark.ui.proxyBase")
	assert.Equal(t, []string{"default", "dev"}, spark.JobNamespaces, "spark.jobNamespaces")
}
