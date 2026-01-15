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

// Package constants defines shared constant values used across the application.
package constants

const (
	// SparkHistoryBase is the base path for Spark History UI.
	SparkHistoryBase = "/history"
	// SparkAppsEndpoint is the Spark History REST endpoint for applications.
	SparkAppsEndpoint = "/api/v1/applications"
	// HealthzURI is the liveness probe endpoint.
	HealthzURI = "/healthz"
	// ReadinessURI is the readiness probe endpoint.
	ReadinessURI = "/readiness"
	// True represents the string value "true".
	True = "true"
)
