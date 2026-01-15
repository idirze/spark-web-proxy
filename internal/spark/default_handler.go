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

// Package spark provides HTTP handlers and proxy wiring for serving Spark UI
// and Spark History endpoints through the reverse proxy.
package spark

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/spark/proxy"
)

// DefaultSparkHandler implements proxy.ReverseProxyHandler for Spark UI and
// Spark History requests.
type DefaultSparkHandler struct {
}

// NewDefaultSparkHandler creates a Spark reverse proxy configured with the
// default request/response rewriting behavior.
func NewDefaultSparkHandler(upstreamURL *url.URL, appID string) *proxy.SparkReverseProxy {
	return proxy.NewSparkReverseProxy(DefaultSparkHandler{}, upstreamURL, appID)
}

// ServeSparkHistory proxies Spark History requests to the configured upstream.
func ServeSparkHistory(c *gin.Context, upstreamURL *url.URL, appID string) {
	NewDefaultSparkHandler(upstreamURL, appID).
		ServeHTTP(c.Writer, c.Request)
}

// ServeSparkUI proxies Spark UI requests to the configured upstream and applies
// Spark UI–specific error handling (for redirects and fallback behavior).
func ServeSparkUI(c *gin.Context, upstreamURL *url.URL, appID string) {
	NewDefaultSparkHandler(upstreamURL, appID).
		WithSparkUIErrorHandler(c.Request.URL).
		ServeHTTP(c.Writer, c.Request)
}

// ModifyRequest returns a function that rewrites the incoming request URL to
// target the provided upstream URL.
func (c DefaultSparkHandler) ModifyRequest(upstreamURL *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = upstreamURL.Scheme
		req.URL.Host = upstreamURL.Host
		req.Host = upstreamURL.Host
		upstreamURL.RawQuery = req.URL.RawQuery
		upstreamURL.RawFragment = req.URL.RawFragment
		req.URL = upstreamURL
	}
}

// ModifyResponse returns a function that rewrites redirect Location headers so
// they remain relative when responses pass through the reverse proxy.
func (c DefaultSparkHandler) ModifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if location == "" {
				log.Warn("No Location header found in the response")
				return nil
			}
			parsedURL, err := url.Parse(location)
			if err != nil {
				log.Error("Error parsing Location URL: %+v", err)
				return nil
			}

			parsedURL.Scheme = ""
			parsedURL.Host = ""

			newLocation := parsedURL.String()
			resp.Header.Set("Location", newLocation)

			log.Debug("Rewritten Location Header: %s", newLocation)
			return nil
		}

		return nil
	}
}
