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

// Package utils provides HTTP-related helper functions used by the application.
package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ValidateURL checks if a given string is a valid URL and panics if it is not.
func ValidateURL(u string, label string) {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		panic(fmt.Sprintf("🚨 %s => %q (Error: %v)", label, u, err))
	}
}

// IsBrowserRequest checks if the incoming HTTP request is from a browser.
// It does so by inspecting the "User-Agent" header and checking for the presence of the word "Mozilla",
// which is part of most modern browser User-Agent strings.
//
// Returns true if the request comes from a browser (i.e., the User-Agent contains "Mozilla"), otherwise false.
func IsBrowserRequest(req *http.Request) bool {
	userAgent := req.Header.Get("User-Agent")
	return strings.Contains(userAgent, "Mozilla")
}
