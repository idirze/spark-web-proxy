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

// Package model defines domain models used across the application.
package model

// SparkAppStatus represents the lifecycle status of a Spark application.
type SparkAppStatus string

const (
	// AppPending indicates that the Spark application is pending execution.
	AppPending SparkAppStatus = "Pending"
	// AppRunning indicates that the Spark application is currently running.
	AppRunning SparkAppStatus = "Running"
	// AppSucceeded indicates that the Spark application has completed successfully.
	AppSucceeded SparkAppStatus = "Succeeded"
	// AppFailed indicates that the Spark application has failed.
	AppFailed SparkAppStatus = "Failed"
	// AppUnknown indicates that the Spark application status is unknown.
	AppUnknown SparkAppStatus = "Unknown"
)
