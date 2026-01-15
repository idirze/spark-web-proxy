/*
 *    Copyright 2026 The OKDP Authors.
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

	"github.com/gin-gonic/gin"

	"github.com/okdp/spark-web-proxy/internal/config"
	sparkclient "github.com/okdp/spark-web-proxy/internal/discovery/resolvers/rest"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
	"github.com/okdp/spark-web-proxy/internal/utils"
)

// SparkAppsController handles requests related to Spark applications.
type SparkAppsController struct {
	sparkHistoryBaseURL string
}

// NewSparkAppsController creates a SparkAppsController using the application configuration.
func NewSparkAppsController(config *config.ApplicationConfig) *SparkAppsController {
	return &SparkAppsController{
		sparkHistoryBaseURL: config.GetSparkHistoryBaseURL(),
	}
}

// HandleIncompleteApplications returns a unified list of Spark applications that are
// currently running or not yet fully available in Spark History.
//
// This handler addresses the gap where Spark applications may be visible in Kubernetes
// (as running pods) but are not yet listed in Spark History because event logs have not
// been persisted (e.g. delayed S3 uploads).
//
// The handler:
//  1. Fetches applications from Spark History Server
//  2. Discovers running Spark applications from Kubernetes
//  3. Queries each running application's Spark UI for live application metadata
//  4. Merges history and live applications into a single list, de-duplicated by app ID
//
// If an application exists in both Spark History and the live runtime, the Spark History
// representation is preferred.
//
// The response format is compatible with the Spark History Server API and can be consumed
// directly by the Spark UI.
func (r SparkAppsController) HandleIncompleteApplications(c *gin.Context) {

	sparkHistoryClient, err := sparkclient.NewSparkRestClient(c.Request, r.sparkHistoryBaseURL)
	if err != nil {
		log.Error("Unable to create new spark history client: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unable to create new spark history client from upstream URL: %s", r.sparkHistoryBaseURL)})
		return
	}

	historyApps, err := sparkHistoryClient.GetApplications()
	if err != nil {
		log.Error("Failed to list spark applications in spark history from upstream URL %s: %+v", r.sparkHistoryBaseURL, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to list spark applications from upstream URL: %s", r.sparkHistoryBaseURL)})
		return
	}

	runningApps := model.GetRunningSparkApps()
	uncompletedApps := make([]model.SparkApp, 0, len(runningApps))
	for _, running := range runningApps {
		sparkClient, err := sparkclient.NewSparkRestClient(c.Request, running.BaseURL)
		if err != nil {
			log.Warn("Unable to create new spark app client: %+v", err)
			continue
		}

		app, err := sparkClient.GetApplicationInfo(running.AppID)
		if err != nil {
			log.Warn("Unable to fetch application info for %s: %v", running.AppID, err)
			continue
		}
		uncompletedApps = append(uncompletedApps, *app)
	}

	running := utils.MergeByKey(*historyApps, uncompletedApps, func(a model.SparkApp) string { return a.ID })

	c.JSON(http.StatusOK, running)
}
