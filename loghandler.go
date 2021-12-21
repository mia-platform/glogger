/*
 * Copyright 2021 Mia srl
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

	"github.com/sirupsen/logrus"
)

type LoggerHandler struct {
	Logger         *logrus.Logger
	ExcludedPrefix []string
	OnNext         func(rrw *ReadableResponseWriter, httpRequest *http.Request)
}

func HandleLog(loggerHandler LoggerHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := getReqID(loggerHandler.Logger, r.Header)
		ctx := WithLogger(r.Context(), logrus.NewEntry(loggerHandler.Logger).WithFields(logrus.Fields{
			"reqId": requestID,
		}))
		myw := ReadableResponseWriter{writer: w, statusCode: http.StatusOK}

		// Skip logging for excluded routes
		for _, prefix := range loggerHandler.ExcludedPrefix {
			if strings.HasPrefix(r.URL.RequestURI(), prefix) {
				loggerHandler.OnNext(&myw, r.WithContext(ctx))
				return
			}
		}

		Get(ctx).WithFields(logrus.Fields{
			"http": HTTP{
				Request: &Request{
					Method:    r.Method,
					UserAgent: map[string]interface{}{"original": r.Header.Get("user-agent")},
				},
			},
			"url": URL{Path: r.URL.RequestURI()},
			"host": Host{
				ForwardedHost: r.Header.Get(forwardedHostHeaderKey),
				Hostname:      removePort(r.Host),
				IP:            r.Header.Get(forwardedForHeaderKey),
			},
		}).Trace("incoming request")

		loggerHandler.OnNext(&myw, r.WithContext(ctx))

		Get(ctx).WithFields(logrus.Fields{
			"http": HTTP{
				Request: &Request{
					Method:    r.Method,
					UserAgent: map[string]interface{}{"original": r.Header.Get("user-agent")},
				},
				Response: &Response{
					StatusCode: myw.statusCode,
					Body: map[string]interface{}{
						"bytes": getBodyLength(myw),
					},
				},
			},
			"url": URL{Path: r.URL.RequestURI()},
			"host": Host{
				ForwardedHost: r.Header.Get(forwardedHostHeaderKey),
				Hostname:      removePort(r.Host),
				IP:            r.Header.Get(forwardedForHeaderKey),
			},
			"responseTime": float64(time.Since(start).Milliseconds()),
		}).Info("request completed")
	}
}
