/*
 * Copyright 2019 Mia srl
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package glogger

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func getReqID(logger *logrus.Logger, headers http.Header) string {
	if requestID := headers.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
	requestID, err := uuid.NewRandom()
	logger.WithField("requestId", requestID).Trace("generated request id")
	if err != nil {
		logger.WithError(err).Fatal("error generating request id")
	}
	return requestID.String()
}

// RequestMiddlewareLogger is a gorilla/mux middleware to log all requests with logrus
// It logs the incoming request and when request is completed, adding latency of the request
func RequestMiddlewareLogger(logger *logrus.Logger, excludedPrefix []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := getReqID(logger, r.Header)
			ctx := WithLogger(r.Context(), logrus.NewEntry(logger).WithFields(logrus.Fields{
				"reqId": requestID,
			}))
			myw := readableResponseWriter{writer: w, statusCode: http.StatusOK}

			// Skip logging for excluded routes
			for _, prefix := range excludedPrefix {
				if strings.HasPrefix(r.URL.RequestURI(), prefix) {
					next.ServeHTTP(&myw, r.WithContext(ctx))
					return
				}
			}

			Get(ctx).WithFields(logrus.Fields{
				"http":	map[string]interface{}{
					"request": map[string]interface{}{
						"method": r.Method,
						"userAgent": map[string]interface{}{
							"original": r.Header.Get("user-agent"),
						},
					},
				},
				"url": map[string]interface{}{
					"path": r.URL.String(),
				},
				"host": map[string]interface{}{
					"hostname": r.URL.Hostname(),
					"ip":  r.Header.Get("x-forwaded-for"),
				},
			}).Info("incoming request")

			next.ServeHTTP(&myw, r.WithContext(ctx))

			Get(ctx).WithFields(logrus.Fields{
				"http":	map[string]interface{}{
					"request": map[string]interface{}{
						"method": r.Method,
						"userAgent": map[string]interface{}{
							"original": r.Header.Get("user-agent"),
						},
					},
					"response": map[string]interface{}{
						"statusCode": myw.statusCode,
						"body": map[string]interface{}{
							"bytes": myw.Length(),
						},
					},
				},
				"url": map[string]interface{}{
					"path": r.URL.String(),
				},
				"host": map[string]interface{}{
					"hostname": r.URL.Hostname(),
					"ip":  r.Header.Get("x-forwaded-for"),
				},
				"responseTime": float64(time.Since(start).Milliseconds()) / 1e3,
			}).Info("request completed")
		})
	}
}
