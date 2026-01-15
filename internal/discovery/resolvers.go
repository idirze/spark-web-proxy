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

// Package discovery provides utilities for discovering Spark applications
// from Kubernetes pods or the Spark History Server.
package discovery

import (
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	sparkclient "github.com/okdp/spark-web-proxy/internal/discovery/resolvers/rest"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
	"github.com/okdp/spark-web-proxy/internal/utils"
)

// ResolveSparkAppFromPod resolves a Spark application instance from a Kubernetes
// driver pod and registers it in the application model.
func ResolveSparkAppFromPod(pod *corev1.Pod) (*model.SparkAppInstance, error) {
	sparkUIURL := fmt.Sprintf("http://%s:%d", pod.Status.PodIP, utils.GetSparkUIPort(pod))
	sparkApp := &model.SparkAppInstance{
		BaseURL:        sparkUIURL,
		PodName:        pod.Name,
		AppID:          utils.GetSparkAppID(pod),
		Namespace:      pod.Namespace,
		Status:         string(pod.Status.Phase),
		StartTimeEpoch: podStartTimeEpoch(pod),
	}

	model.AddOrUpdateSparkApp(sparkApp)

	return sparkApp, nil
}

// ResolveSparkAppFromHistory resolves a Spark application instance using the
// Spark History Server REST API.
func ResolveSparkAppFromHistory(request *http.Request, sparkHistoryBaseURL string, appID string) (*model.SparkAppInstance, error) {
	sparkClient, err := sparkclient.NewSparkRestClient(request, sparkHistoryBaseURL)
	if err != nil {
		log.Error("Unable to create new spark history client: %+v", err)
		return nil, err
	}
	appInfo, err := sparkClient.GetApplicationInfo(appID)
	if err != nil {
		log.Error("Unable to get spark application '%s' status from spark history, %+v", appID, err)
		return &model.SparkAppInstance{
			Status: string(model.AppUnknown),
		}, err
	}
	sparkAppEnv, err := sparkClient.GetEnvironment(appID)
	if err != nil {
		log.Error("Failed to get the application '%s' environment properties from spark history: %+v", appID, err)
		return &model.SparkAppInstance{
			Status: string(model.AppUnknown),
		}, err
	}
	sparkDriverHost, _ := sparkAppEnv.GetProperty("spark.driver.host")
	sparkDriverPort, _ := sparkAppEnv.GetProperty("spark.ui.port")
	sparkAppID, _ := sparkAppEnv.GetProperty("spark.app.id")
	sparkAppName, _ := sparkAppEnv.GetProperty("spark.app.name")
	sparkAppNamespace, _ := sparkAppEnv.GetProperty("spark.kubernetes.namespace")
	sparkUIBaseURL := fmt.Sprintf("http://%s:%s", sparkDriverHost, sparkDriverPort)

	sparkApp := &model.SparkAppInstance{
		BaseURL:   sparkUIBaseURL,
		PodName:   sparkAppName,
		AppID:     sparkAppID,
		Namespace: sparkAppNamespace,
		Status:    string(model.AppUnknown),
	}

	if appInfo.IsRunning() {
		sparkApp.Status = string(model.AppRunning)
	} else {
		model.AddOrUpdateSparkApp(sparkApp)
	}
	return sparkApp, err
}

// podStartTimeEpoch returns the pod start time as a Unix epoch timestamp
// in milliseconds, or -1 if the start time is not available.
func podStartTimeEpoch(pod *corev1.Pod) int64 {
	if pod.Status.StartTime != nil {
		return pod.Status.StartTime.UnixMilli()
	}
	return -1
}
