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

package utils

import (
	"net/http"
	"testing"
)

// TestValidateURL tests the ValidateURL function to ensure it panics on invalid URLs
func TestValidateURL(t *testing.T) {
	// Given
	validURLs := []string{
		"https://example.com",
		"http://example.com",
		"ftp://example.com",
		"https://sub.example.com/path?query=1",
	}

	invalidURLs := []string{
		"invalid-url",
		"htp:/wrong.url",
		"",
		"example.com",
		"https://",
	}

	// Test valid URLs (should NOT panic)
	for _, url := range validURLs {
		t.Run("Valid: "+url, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ValidateURL(%q) panicked unexpectedly: %v", url, r)
				}
			}()
			ValidateURL(url, "The URL is not valid")
		})
	}

	// Test invalid URLs
	for _, url := range invalidURLs {
		t.Run("Invalid: "+url, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("ValidateURL(%q) did NOT panic as expected", url)
				}
			}()
			ValidateURL(url, "The URL is not valid")
		})
	}
}

// TestIsBrowserRequest tests the isBrowserRequest function.
func TestIsBrowserRequest(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		isBrowser bool
	}{
		{
			name:      "Browser request (Chrome)",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			isBrowser: true,
		},
		{
			name:      "Browser request (Firefox)",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			isBrowser: true,
		},
		{
			name:      "Non-browser request (curl)",
			userAgent: "curl/7.68.0",
			isBrowser: false,
		},
		{
			name:      "Non-browser request (Postman)",
			userAgent: "PostmanRuntime/7.28.0",
			isBrowser: false,
		},
		{
			name:      "Empty User-Agent",
			userAgent: "",
			isBrowser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the given User-Agent
			req := &http.Request{
				Header: http.Header{
					"User-Agent": []string{tt.userAgent},
				},
			}

			// Check if the function returns the expected result
			got := IsBrowserRequest(req)
			if got != tt.isBrowser {
				t.Errorf("isBrowserRequest() = %v, want %v", got, tt.isBrowser)
			}
		})
	}
}
