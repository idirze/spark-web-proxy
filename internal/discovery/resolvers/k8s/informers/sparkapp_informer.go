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

// Package informers provides Kubernetes informers used to discover and track
// Spark driver pods and their associated application metadata.
package informers

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/okdp/spark-web-proxy/internal/config"
	"github.com/okdp/spark-web-proxy/internal/discovery"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/model"
)

// SparkAppInformer watches Kubernetes namespaces for Spark driver pods and
// maintains an in-memory view of running Spark applications.
type SparkAppInformer struct {
	namespaces []string
	ui         config.UI
}

// NewSparkAppInformer creates a SparkAppInformer using the application configuration.
func NewSparkAppInformer(config *config.ApplicationConfig) *SparkAppInformer {
	return &SparkAppInformer{
		namespaces: config.Spark.JobNamespaces,
		ui:         config.Spark.UI,
	}
}

// WatchSparkApps starts watching Spark driver pods in all configured namespaces.
func (i SparkAppInformer) WatchSparkApps(clientset *kubernetes.Clientset) {
	namespaces := i.namespaces
	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
	}

	for _, ns := range namespaces {
		go i.WatchNamespaceSparkApps(clientset, ns)
	}
}

// WatchNamespaceSparkApps starts a Spark driver pod informer for a single namespace.
func (i SparkAppInformer) WatchNamespaceSparkApps(clientset *kubernetes.Clientset, namespace string) {

	log.Info("Running spark app informer on the following namespaces: %s", func() string {
		if namespace == metav1.NamespaceAll {
			return "all"
		}
		return namespace
	}())

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Minute,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			opts.LabelSelector = "spark-role=driver"
		}),
	)

	podInformer := factory.Core().V1().Pods().Informer()

	// Register event handlers
	registration, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: i.sparkAppAddedOrUpdated,
		UpdateFunc: func(_, newObj interface{}) {
			i.sparkAppAddedOrUpdated(newObj)
		},
		DeleteFunc: i.sparkAppDeleted,
	})

	if err != nil {
		log.Error("Failed to add spark app event handler: %+v", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	defer cancel()

	factory.Start(ctx.Done())

	<-ctx.Done()

	log.Info("Received shutdown signal. Stopping Spark app informer...")
	_ = podInformer.RemoveEventHandler(registration)
	log.Info("Spark app informer successfully stopped.")
}

func (i SparkAppInformer) sparkAppAddedOrUpdated(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	sparkApp, _ := discovery.ResolveSparkAppFromPod(pod)
	log.Info("The application '%s' (%s/%s) was updated: %s at %s", sparkApp.AppID, sparkApp.Namespace, pod.Name, sparkApp.Status, sparkApp.BaseURL)
}

func (i SparkAppInformer) sparkAppDeleted(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	sparkApp, _ := model.DeleteSparkAppByName(pod.Name)
	log.Info("The application '%s' (%s/%s) was removed", sparkApp.AppID, pod.Namespace, pod.Name)
}
