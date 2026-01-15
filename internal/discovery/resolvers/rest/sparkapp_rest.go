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

// Package sparkclient provides a REST client for interacting with the Spark
// History Server API.
package sparkclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/okdp/spark-web-proxy/internal/constants"
	restclient "github.com/okdp/spark-web-proxy/internal/discovery/resolvers/rest/client"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
)

// SparkRestClient provides high-level methods to query the Spark History Server API.
type SparkRestClient struct {
	*restclient.SparkClient
}

// NewSparkRestClient creates a SparkRestClient for forwarding an incoming HTTP request
// to the Spark History Server API.
func NewSparkRestClient(request *http.Request, sparkHistoryBaseURL string) (*SparkRestClient, error) {
	client, err := restclient.NewSparkClient(request, sparkHistoryBaseURL)
	return &SparkRestClient{
		client,
	}, err
}

// GetApplications retrieves the list of applications from the Spark History Server.
func (c *SparkRestClient) GetApplications() (*[]model.SparkApp, error) {

	log.Debug("Get the list of spark history applications from URL: %s", c.Request.URL.String())

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		return nil, err
	}

	return doResponse[[]model.SparkApp](resp, "")
}

// GetApplicationInfo retrieves application details for the given application ID.
func (c *SparkRestClient) GetApplicationInfo(appID string) (*model.SparkApp, error) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s", constants.SparkAppsEndpoint, appID)

	log.Debug("Get the application status '%s' from URL: %s", appID, c.Request.URL.String())

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		return nil, err
	}

	return doResponse[model.SparkApp](resp, appID)
}

// GetEnvironment retrieves environment properties for the given application ID.
func (c *SparkRestClient) GetEnvironment(appID string) (*model.SparkAppEnvironment, error) {
	c.Request.URL.Path = fmt.Sprintf("%s/%s/%s", constants.SparkAppsEndpoint, appID, "environment")

	log.Debug("Get the application '%s' environment properties from URL: %s", appID, c.Request.URL.String())

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		return nil, err
	}

	return doResponse[model.SparkAppEnvironment](resp, appID)
}

// doResponse validates that the upstream response contains JSON and decodes it into T.
func doResponse[T any](response *http.Response, appID string) (*T, error) {
	var object T
	ct := strings.ToLower(response.Header.Get("Content-Type"))

	log.Debug("Upstream response: status=%d content-encoding=%q content-type=%q content-length=%q",
		response.StatusCode, response.Header.Get("Content-Encoding"), response.Header.Get("Content-Type"), response.Header.Get("Content-Length"),
	)

	// Not JSON content-type: log snippet and fail fast
	if !strings.Contains(ct, "application/json") && !strings.Contains(ct, "text/json") {
		return nil, fmt.Errorf("spark UI is initializing")
	}

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		log.Error("Failed to decode JSON for '%s': %v", appID, err)
		return nil, err
	}

	return &object, nil

}
