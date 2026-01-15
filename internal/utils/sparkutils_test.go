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

//revive:disable
package utils

import (
	"testing"
)

func TestCleanURLPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			// Test with a valid URL containing '/kill' and query parameters
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/jobs/job/kill?id=0",
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/jobs",
		},
		{
			// Test with a URL containing '/kill' with query parameters and no extra path after it
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/stages/stage/kill?id=0",
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/stages",
		},
		{
			// Test with a valid URL containing '/kill' and query parameters
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/jobs/job/kill/?id=2",
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/jobs",
		},
		{
			// Test with a URL containing '/kill' with query parameters and no extra path after it
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/stages/stage/kill/?id=3",
			"/sparkui/spark-4feb1501874842e8854dae05e4e81b19/stages",
		},
		{
			// Test with a valid URL but no '/kill' present (no change expected)
			"/spark-4feb1501874842e8854dae05e4e81b19/jobs",
			"/spark-4feb1501874842e8854dae05e4e81b19/jobs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanKillURLPath(tt.input)
			if got != tt.expected {
				t.Errorf("cleanURLPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
