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

// Package controllers provides HTTP handlers for Spark Web Proxy endpoints.
package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/okdp/spark-web-proxy/internal/config"
	"github.com/okdp/spark-web-proxy/internal/constants"
	"github.com/okdp/spark-web-proxy/internal/discovery"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
	"github.com/okdp/spark-web-proxy/internal/spark"
)

// SparkHistoryController handles requests that are routed to the Spark History Server
// and manages redirects to the Spark UI when an application is still running.
type SparkHistoryController struct {
	sparkHistoryBaseURL string
	sparkHistoryBase    string
	sparkUIProxyBase    string
}

// NewSparkHistoryController creates a SparkHistoryController using the application configuration.
func NewSparkHistoryController(config *config.ApplicationConfig) *SparkHistoryController {
	controller := &SparkHistoryController{
		sparkHistoryBaseURL: config.GetSparkHistoryBaseURL(),
		sparkHistoryBase:    constants.SparkHistoryBase,
		sparkUIProxyBase:    strings.TrimSpace(config.Spark.UI.ProxyBase),
	}

	log.Info("Spark History K8S Service URL: %s, Spark UI Proxy base: %s", controller.sparkHistoryBaseURL, controller.sparkUIProxyBase)
	return controller
}

// HandleHistoryApp handles Spark History application routes (e.g. /history/:appID/*path).
// If the application is still running, it redirects to the Spark UI; otherwise it
// proxies the request to the Spark History Server.
func (r SparkHistoryController) HandleHistoryApp(c *gin.Context) {

	appID := c.Param("appID")
	jobPath := c.Param("path")

	sparkApp, found := model.GetSparkApp(appID)

	// The application was started in cluster mode and is running
	if found && sparkApp.IsRunning() {
		r.redirectToSparkUI(c, appID)
		return
	}

	// The application was started in client or cluster mode and was not present locally
	if !found {
		log.Debug("The application '%s' was not found locally, checking in spark history ...", appID)
		sparkApp, _ := discovery.ResolveSparkAppFromHistory(c.Request, r.sparkHistoryBaseURL, appID)
		if sparkApp.IsRunning() {
			r.redirectToSparkUI(c, appID)
			return
		}
	}

	log.Debug("The application '%s' was started in client or cluster mode and is completed, forward to spark history: %s", appID, r.sparkHistoryBaseURL)

	upstreamURL, err := url.Parse(fmt.Sprintf("%s%s/%s%s", r.sparkHistoryBaseURL, r.sparkHistoryBase, appID, jobPath))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid upstream URL: %s", upstreamURL)})
		return
	}

	spark.ServeSparkHistory(c, upstreamURL, appID)
}

// HandleDefault proxies non-application Spark History routes to the Spark History Server.
func (r SparkHistoryController) HandleDefault(c *gin.Context) {
	r.serveSparkHistory(c, spark.ServeSparkHistory)
}

// HandleIncompleteApps proxies Spark History routes and injects content for the
// "incomplete applications" pages when applicable.
func (r SparkHistoryController) HandleIncompleteApps(c *gin.Context) {
	r.serveSparkHistory(c, spark.ServeSparkHistoryIncompleteApps)
}

// serveSparkHistory proxies the current request path to the Spark History Server
// using the provided serve function.
func (r SparkHistoryController) serveSparkHistory(c *gin.Context, serve func(*gin.Context, *url.URL, string)) {
	path := c.Request.URL.Path

	upstreamURL, err := url.Parse(r.sparkHistoryBaseURL + path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid upstream URL: %s", r.sparkHistoryBaseURL+path),
		})
		return
	}

	serve(c, upstreamURL, "")
}

// redirectToSparkUI redirects the client to the proxied Spark UI jobs page for the given app ID.
func (r SparkHistoryController) redirectToSparkUI(c *gin.Context, appID string) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s/jobs/", r.sparkUIProxyBase, appID)
	log.Debug("The application '%s' is running, redirect to spark ui '%s'", appID, c.Request.URL.String())
	c.Redirect(http.StatusFound, c.Request.URL.String())
}
