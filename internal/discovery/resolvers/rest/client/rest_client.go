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

// Package sparkclient provides a lightweight HTTP client for interacting
// with the Spark History Server APIs.
package sparkclient

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/okdp/spark-web-proxy/internal/constants"
	log "github.com/okdp/spark-web-proxy/internal/logging"
)

// SparkClient wraps an HTTP client and request used to communicate
// with the Spark History Server.
type SparkClient struct {
	Client  *http.Client
	Request *http.Request
}

// NewSparkClient creates a new SparkClient for forwarding an incoming HTTP
// request to the Spark History Server applications API.
func NewSparkClient(request *http.Request, sparkHistoryBaseURL string) (*SparkClient, error) {
	jar, _ := cookiejar.New(nil)
	apiURL := fmt.Sprintf("%s%s", sparkHistoryBaseURL, constants.SparkAppsEndpoint)
	req, err := http.NewRequest(request.Method, apiURL, nil)
	if err != nil {
		log.Error("failed to create request: %+v", err)
		return nil, err
	}

	// Forward query parameters
	req.URL.RawQuery = request.URL.RawQuery

	// Copy headers from the original request
	for key, values := range request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	// Copy cookies from the original request
	for _, cookie := range request.Cookies() {
		req.AddCookie(cookie)
	}

	// Disable compression for API calls
	req.Header.Del("Accept-Encoding")
	req.Header.Set("Accept-Encoding", "identity")

	return &SparkClient{
		Client:  &http.Client{Jar: jar},
		Request: req,
	}, nil
}
