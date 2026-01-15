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

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// SparkReverseProxy wraps httputil.ReverseProxy and adds Spark-specific
// context such as the application ID.
type SparkReverseProxy struct {
	*httputil.ReverseProxy
	appID string
}

// NewSparkReverseProxy creates a new SparkReverseProxy configured with
// request and response modifiers and a default error handler.
func NewSparkReverseProxy(c ReverseProxyHandler, upstreamURL *url.URL, appID string) *SparkReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	proxy.Director = c.ModifyRequest(upstreamURL)
	proxy.ModifyResponse = c.ModifyResponse()
	proxy.ErrorHandler = DefaultErrorHandler(appID)
	return &SparkReverseProxy{proxy, appID}
}

// WithSparkUIErrorHandler configures the proxy to use a Spark UI–specific
// error handler and returns the updated proxy.
func (p *SparkReverseProxy) WithSparkUIErrorHandler(fromURL *url.URL) *SparkReverseProxy {
	p.ErrorHandler = SparkUIErrorHandler(fromURL, p.appID)
	return p
}

// ServeHTTP implements http.Handler by delegating the request handling
// to the underlying ReverseProxy.
func (p *SparkReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ReverseProxy.ServeHTTP(w, r)
}
