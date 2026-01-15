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

// Package model defines domain models used across the application.
package model

import (
	"sync"
)

// SparkAppInstance represents a running or completed Spark application
// discovered either from Kubernetes or the Spark History Server.
type SparkAppInstance struct {
	BaseURL        string
	PodName        string
	AppID          string
	Namespace      string
	Status         string
	StartTimeEpoch int64
}

// SparkAppsStore holds a concurrent map of Spark applications, keyed by appId.
var (
	SparkAppsStore = struct {
		Instances sync.Map
	}{}
)

// IsRunning reports whether the Spark application is currently running.
func (app SparkAppInstance) IsRunning() bool {
	return app.Status == string(AppRunning)
}

// IsCompleted reports whether the Spark application has completed
// (i.e., it is not in the running state).
func (app SparkAppInstance) IsCompleted() bool {
	return !app.IsRunning()
}

// AddOrUpdateSparkApp adds a new SparkApp to the map or updates an existing one
func AddOrUpdateSparkApp(app *SparkAppInstance) {
	SparkAppsStore.Instances.Store(app.AppID, app)
}

// MakeSparkAppCompleted updates SparkApp to AppUnknown status
func MakeSparkAppCompleted(appID string) {
	app, found := GetSparkApp(appID)
	if found {
		app.Status = string(AppUnknown)
	} else {
		app = &SparkAppInstance{
			AppID:  appID,
			Status: string(AppUnknown),
		}
	}

	SparkAppsStore.Instances.Store(appID, app)
}

// DeleteSparkApp removes a SparkApp from the map
func DeleteSparkApp(appID string) {
	SparkAppsStore.Instances.Delete(appID)
}

// DeleteSparkAppByName removes a Spark application from the map by its PodName
// and returns the deleted SparkApp.
//
// It iterates over the sync.Map, finds the matching SparkApp by PodName,
// deletes the entry, and returns the deleted SparkApp.
//
// Parameters:
//   - podName: The name of the pod to be removed.
//
// Returns:
//   - (SparkApp, bool): The deleted SparkApp and a boolean indicating success.
//
// Example usage:
//
//	deletedApp, found := DeleteSparkAppByPodName("spark-pod-123")
//	if found {
//	    fmt.Println("Deleted SparkApp:", deletedApp)
//	}
func DeleteSparkAppByName(podName string) (*SparkAppInstance, bool) {
	var deletedApp *SparkAppInstance
	var found bool

	SparkAppsStore.Instances.Range(func(key, value interface{}) bool {
		if app, ok := value.(*SparkAppInstance); ok && app.PodName == podName {
			deletedApp = app
			found = true
			SparkAppsStore.Instances.Delete(key)
			return false
		}
		return true
	})

	return deletedApp, found
}

// GetSparkApp retrieves a SparkApp from the map by appID
func GetSparkApp(appID string) (*SparkAppInstance, bool) {
	value, exists := SparkAppsStore.Instances.Load(appID)
	if exists {
		return value.(*SparkAppInstance), exists
	}
	return &SparkAppInstance{}, false
}

// GetRunningSparkApps retrieves all SparkApps from the map
func GetRunningSparkApps() []*SparkAppInstance {
	apps := make([]*SparkAppInstance, 0)
	SparkAppsStore.Instances.Range(func(_, value interface{}) bool {
		app := value.(*SparkAppInstance)
		if app != nil && app.IsRunning() {
			apps = append(apps, app)
		}
		return true
	})
	return apps
}

// GetProperty retrieves the value for the specified property name from the SparkProperties slice.
// It returns the value as a string and a boolean indicating whether the property was found.
//
// Parameters:
//   - propertyName (string): The name of the property to retrieve.
//
// Returns:
//   - (string, bool): The value of the property if found, and true. If the property is not found, it returns an empty string and false.
//
// Example:
//
//	response := SparkHistoryEnvironmentResponse{
//	    SparkProperties: [][]string{
//	        {"spark.acls.enable", "true"},
//	        {"spark.app.id", "spark-xyz123"},
//	    },
//	}
//	value, found := response.GetProperty("spark.app.id")
//	fmt.Println(value, found) // Output: "spark-xyz123 true"
func (app SparkAppEnvironment) GetProperty(propertyName string) (string, bool) {
	for _, property := range app.SparkProperties {
		if property[0] == propertyName {
			return property[1], true
		}
	}
	return "_", false
}
