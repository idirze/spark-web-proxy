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

// Package proxy provides reverse-proxy helpers and error handlers used
// to route requests to Spark UIs and Spark History.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
	"github.com/okdp/spark-web-proxy/internal/utils"
)

// ReverseProxyHandler defines hooks used to customize the reverse proxy
// request and response processing.
type ReverseProxyHandler interface {
	ModifyRequest(upstreamURL *url.URL) func(*http.Request)
	ModifyResponse() func(*http.Response) error
}

// DefaultErrorHandler returns a function that handles errors by logging the
// error details and sending an HTTP 502 (Bad Gateway) response with the error message.
//
// This error handler is typically used to handle proxy errors in situations where
// a service behind the proxy returns an unexpected error. The function logs the
// error with the URL path and details, and then responds with a standardized
// message to the client, indicating that a proxy error occurred.
//
// Parameters:
//   - rw: The `http.ResponseWriter` used to write the error response to the client.
//   - req: The incoming `http.Request` containing the original request details.
//   - err: The error that occurred during the request processing.
//
// Returns:
//   - A function of type `func(http.ResponseWriter, *http.Request, error)` that handles
//     the error and sends the appropriate response back to the client.
func DefaultErrorHandler(appID string) func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, req *http.Request, err error) {
		if isCancelErr(err) {
			log.Debug("Request canceled for app '%s' url=%s: %v", appID, req.URL.String(), err)
			return
		}
		log.Error("An error was occured when accessing the application '%s' at URL: %s, \ndetails: %+v", appID, req.URL.String(), err)
		http.Error(rw, fmt.Sprintf("An error was occured when accessing the application '%s' at URL: %s, %s", appID, req.URL.String(), err.Error()), http.StatusBadGateway)
	}
}

// SparkUIErrorHandler returns an error handler tailored for Spark UI requests.
// It handles expected cancellation errors quietly, supports browser redirects
// for kill actions, and falls back to Spark History when the Spark UI becomes
// unavailable.
func SparkUIErrorHandler(fromURL *url.URL, appID string) func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, req *http.Request, err error) {
		if isCancelErr(err) {
			log.Debug("Request canceled for app '%s' url=%s: %v", appID, req.URL.String(), err)
			return
		}
		if strings.Contains(fromURL.Path, "/kill") && utils.IsBrowserRequest(req) {
			previousPage := utils.CleanKillURLPath(fromURL.Path)
			log.Info("A spark job or stage kill was received '%s' for application '%s', redirecting to previous page: %s", fromURL.Path, appID, previousPage)
			rw.Header().Set("Location", previousPage)
			rw.WriteHeader(http.StatusFound)
			return
		}
		log.Error("An error was occured when accessing spark application '%s' at URL: %s, redirect to spark history \ndetails: %+v", appID, req.URL.String(), err)
		model.MakeSparkAppCompleted(appID)
		// redirect to spark history
		rw.Header().Set("Location", fromURL.Path)
		rw.WriteHeader(http.StatusFound)
	}
}

// isCancelErr reports whether the given error is caused by a request
// cancellation or timeout rather than a real upstream failure.
//
// This typically happens when:
//   - the client (browser) closes the connection
//   - the user navigates away or refreshes the page
//   - the request context is canceled by Gin / net/http
//   - a request times out while waiting for a response
//
// These errors are expected in reverse proxies and should usually be
// logged at debug or warn level, not as hard errors.
func isCancelErr(err error) bool {
	if err == nil {
		return false
	}

	// Context canceled by client disconnect or request abort
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Network timeout (common when client disconnects mid-stream)
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}

	return false
}
