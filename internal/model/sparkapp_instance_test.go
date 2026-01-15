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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProperty(t *testing.T) {
	// Define a sample SparkHistoryEnvironmentResponse
	response := SparkAppEnvironment{
		SparkProperties: [][]string{
			{"spark.acls.enable", "true"},
			{"spark.app.id", "spark-9adcb03756d042de8f2f5c7deb8715b3"},
			{"spark.ui.filters", "io.okdp.spark.authc.OidcAuthFilter"},
		},
	}

	// Test: Property exists (spark.app.id)
	t.Run("Property exists", func(t *testing.T) {
		value, found := response.GetProperty("spark.app.id")
		assert.True(t, found, "Property should be found")
		assert.Equal(t, "spark-9adcb03756d042de8f2f5c7deb8715b3", value, "The value should match")
	})

	// Test: Property exists (spark.acls.enable)
	t.Run("Property exists (acls)", func(t *testing.T) {
		value, found := response.GetProperty("spark.acls.enable")
		assert.True(t, found, "Property should be found")
		assert.Equal(t, "true", value, "The value should match")
	})

	// Test: Property does not exist (non-existing property)
	t.Run("Property does not exist", func(t *testing.T) {
		value, found := response.GetProperty("spark.non.existing.property")
		assert.False(t, found, "Property should not be found")
		assert.Equal(t, "_", value, "The value should be _")
	})
}
