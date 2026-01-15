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

// Package security provides HTTP security middleware such as CORS and
// custom header configuration for the Gin web framework.
package security

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/okdp/spark-web-proxy/internal/config"
)

// HTTPSecurity returns a list of Gin middleware handlers that apply
// security-related settings such as CORS and custom HTTP headers.
func HTTPSecurity(securityConfig config.Security) []gin.HandlerFunc {
	var handlers = []gin.HandlerFunc{}
	return append(handlers, Cors(securityConfig.Cors), Headers(securityConfig.Headers))
}

// Cors creates a Gin middleware handler that applies CORS configuration
// based on the provided application settings.
func Cors(corsConfig config.Cors) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     corsConfig.AllowedOrigins,
		AllowMethods:     corsConfig.AllowedMethods,
		AllowHeaders:     corsConfig.AllowedHeaders,
		ExposeHeaders:    corsConfig.ExposedHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           time.Duration(corsConfig.MaxAge) * time.Second,
	})
}

// Headers creates a Gin middleware handler that injects custom HTTP
// headers into every response.
func Headers(headersConf map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for header, value := range headersConf {
			c.Header(header, value)
		}
		c.Next()
	}
}
