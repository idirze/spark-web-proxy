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

// SparkApp represents the structure of a Spark application info JSON response.
// SparkApp represents a Spark application as returned by the Spark
// History Server REST API.
type SparkApp struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Attempts []SparkAppAttempt `json:"attempts,omitempty"`
}

// SparkAppAttempt represents a single execution attempt of a Spark application.
type SparkAppAttempt struct {
	StartTime        string `json:"startTime,omitempty"`
	EndTime          string `json:"endTime,omitempty"`
	LastUpdated      string `json:"lastUpdated,omitempty"`
	Duration         int64  `json:"duration,omitempty"`
	SparkUser        string `json:"sparkUser,omitempty"`
	Completed        bool   `json:"completed,omitempty"`
	AppSparkVersion  string `json:"appSparkVersion,omitempty"`
	StartTimeEpoch   int64  `json:"startTimeEpoch,omitempty"`
	EndTimeEpoch     int64  `json:"endTimeEpoch,omitempty"`
	LastUpdatedEpoch int64  `json:"lastUpdatedEpoch,omitempty"`
}

// SparkAppEnvironment represents the JSON structure for Spark history environment response (/applications/[app-id]/environment).
type SparkAppEnvironment struct {
	SparkProperties [][]string `json:"sparkProperties"`
}

// IsRunning checks if the Spark application is still running.
// It returns true if at least one attempt meets any of the following conditions:
// 1. Completed is false
// 2. Duration is 0
// 3. EndTimeEpoch is -1
func (app SparkApp) IsRunning() bool {
	for _, attempt := range app.Attempts {
		if !attempt.Completed ||
			attempt.Duration == 0 ||
			attempt.EndTimeEpoch == -1 {
			return true
		}
	}
	return false
}
