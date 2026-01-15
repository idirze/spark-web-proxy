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

// Package utils provides Spark related helper functions used by the application.
package utils

import (
	"regexp"
	"time"
)

// CleanKillURLPath cleans the Spark URL job or stage kill path
func CleanKillURLPath(path string) string {
	re := regexp.MustCompile(`(.*)/[^/]+/kill[/]{0,1}(\?.*)?$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return path
}

// FormatSparkTime converts an epoch timestamp in milliseconds to the
// Spark History Server time format.
//
// The returned format matches Spark's UI and REST API expectations:
//
//	YYYY-MM-DDTHH:mm:ss.SSSGMT
//
// Example:
//
//	epochMillis: 1767710303938
//	output:      "2026-01-06T14:38:23.938GMT"
//
// The time is always formatted in UTC (GMT).
func FormatSparkTime(epochMillis int64) string {
	return time.UnixMilli(epochMillis).
		UTC().
		Format("2006-01-02T15:04:05.000GMT")
}

// MergeByKey merges two slices of the same type into one.
// If an element exists in both slices (same key), the element from `preferred` wins.
// Memory notes:
// - Allocates exactly one output slice with capacity len(preferred)+len(other)
// - Allocates one map sized to len(other) (only items inserted from `other`)
func MergeByKey[T any](preferred []T, other []T, keyFn func(T) string) []T {

	merged := make([]T, 0, len(preferred)+len(other))
	index := make(map[string]int, len(other))

	for _, v := range other {
		k := keyFn(v)
		if k == "" {
			continue
		}
		index[k] = len(merged)
		merged = append(merged, v)
	}

	for _, v := range preferred {
		k := keyFn(v)
		if k == "" {
			continue
		}
		if i, ok := index[k]; ok {
			merged[i] = v
			continue
		}
		merged = append(merged, v)
	}

	return merged
}
