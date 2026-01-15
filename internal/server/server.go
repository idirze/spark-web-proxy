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

// Package server provides the HTTP server setup for the Spark Web Proxy.
package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/okdp/spark-web-proxy/internal/config"
	"github.com/okdp/spark-web-proxy/internal/constants"
	"github.com/okdp/spark-web-proxy/internal/controllers"
	"github.com/okdp/spark-web-proxy/internal/discovery/resolvers/k8s/informers"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/security"
)

// NewSparkUIProxyServer creates and configures the HTTP server for the Spark Web Proxy.
// It initializes Kubernetes informers, configures the Gin router, and registers routes
// for Spark UI, Spark History, and health endpoints.
func NewSparkUIProxyServer(config *config.ApplicationConfig) *http.Server {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Failed to load Kubernetes in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatal("Failed to create Kubernetes client: %v", err)
	}

	informer := informers.NewSparkAppInformer(config)

	go informer.WatchSparkApps(clientset)

	// Set up Gin router
	gin.SetMode(config.Proxy.Mode)
	r := gin.New()
	r.Use(log.Logger()...)
	r.Use(gin.Recovery())

	// Apply http security (cors, headers, etc)
	r.Use(security.HTTPSecurity(config.Security)...)

	// Spark UI
	sparkUI := controllers.NewSparkUIController(config)
	sparkHistory := controllers.NewSparkHistoryController(config)
	sparkApps := controllers.NewSparkAppsController(config)

	// Spark UI Handler
	r.Any(fmt.Sprintf("%s/:appID/*path", config.Spark.UI.ProxyBase), sparkUI.HandleRunningApp)

	// Spark history Handlers
	r.Any("/history/:appID/*path", sparkHistory.HandleHistoryApp)
	r.Any("/static/*path", sparkHistory.HandleDefault)
	r.Any("/api/v1/applications", func(c *gin.Context) {
		if c.Query("status") == "running" {
			sparkApps.HandleIncompleteApplications(c)
			return
		}
		sparkHistory.HandleDefault(c)
	})
	r.Any("/api/v1/applications/*path", sparkHistory.HandleDefault)
	r.Any("/history/", sparkHistory.HandleDefault)
	r.Any("/home/", func(c *gin.Context) {
		if c.Query("showIncomplete") == constants.True {
			sparkHistory.HandleIncompleteApps(c)
			return
		}
		sparkHistory.HandleDefault(c)
	})
	r.Any("/jobs/", func(c *gin.Context) {
		if c.Query("showIncomplete") == constants.True {
			sparkHistory.HandleIncompleteApps(c)
			return
		}
		sparkHistory.HandleDefault(c)
	})
	r.Any("/", func(c *gin.Context) {
		if c.Query("showIncomplete") == constants.True {
			sparkHistory.HandleIncompleteApps(c)
			return
		}
		sparkHistory.HandleDefault(c)
	})

	r.GET(constants.HealthzURI, controllers.Healthz)
	r.GET(constants.ReadinessURI, controllers.Readiness)

	proxy := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", config.Proxy.ListenAddress, config.Proxy.Port),
	}

	return proxy
}
