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

// SparkUIController handles requests routed to running Spark application UIs
// and redirects completed applications to Spark History.
type SparkUIController struct {
	sparkHistoryBaseURL string
	sparkHistoryBase    string
	sparkUIProxyBase    string
}

// NewSparkUIController creates a SparkUIController using the application configuration.
func NewSparkUIController(config *config.ApplicationConfig) *SparkUIController {
	return &SparkUIController{
		sparkHistoryBaseURL: config.GetSparkHistoryBaseURL(),
		sparkHistoryBase:    constants.SparkHistoryBase,
		sparkUIProxyBase:    strings.TrimSpace(config.Spark.UI.ProxyBase),
	}
}

// HandleRunningApp handles Spark UI routes for running applications.
// If the application is completed, the request is redirected to Spark History;
// otherwise, it is proxied to the live Spark UI.
func (r SparkUIController) HandleRunningApp(c *gin.Context) {
	appID := c.Param("appID")
	sparkAppPath := strings.TrimPrefix(c.Param("path"), "/")

	sparkApp, found := model.GetSparkApp(appID)

	// The application was started in cluster or client mode and was completed
	if found && sparkApp.IsCompleted() {
		r.redirectToSparkHistory(c, appID)
		return
	}

	// The application was started in client or cluster mode and was not present locally
	if !found {
		log.Debug("The application '%s' was not found locally, checking in spark history ...", appID)
		sparkApp, _ = discovery.ResolveSparkAppFromHistory(c.Request, r.sparkHistoryBaseURL, appID)
		if sparkApp.IsCompleted() {
			r.redirectToSparkHistory(c, appID)
			return
		}
	}

	sparkkUI := fmt.Sprintf("%s/%s", sparkApp.BaseURL, sparkAppPath)
	upstreamURL, err := url.Parse(sparkkUI)
	if err != nil {
		log.Error("Invalid spark ui URL '%s' for the application '%s', redirect to spark history", sparkkUI, appID)
		model.MakeSparkAppCompleted(appID)
		r.redirectToSparkHistory(c, appID)
		return
	}

	if r.sparkUIProxyBase != "/proxy" {
		sparkUIRoot := fmt.Sprintf("%s/%s", r.sparkUIProxyBase, appID)
		c.Request.Header.Add("X-Forwarded-Context", sparkUIRoot)
	}

	spark.ServeSparkUI(c, upstreamURL, appID)
}

// redirectToSparkHistory redirects the client to the Spark History page
// for the given application ID.
func (r SparkUIController) redirectToSparkHistory(c *gin.Context, appID string) {
	c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, r.sparkUIProxyBase, r.sparkHistoryBase)
	log.Debug("The application '%s' was completed, redirect to spark history '%s'", appID, c.Request.URL.String())
	c.Redirect(http.StatusFound, c.Request.URL.String())
}
