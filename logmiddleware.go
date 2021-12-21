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
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	forwardedHostHeaderKey = "x-forwarded-host"
	forwardedForHeaderKey  = "x-forwarded-for"
)

// HTTP is the struct of the log formatter.
type HTTP struct {
	Request  *Request  `json:"request,omitempty"`
	Response *Response `json:"response,omitempty"`
}

// Request contains the items of request info log.
type Request struct {
	Method    string                 `json:"method,omitempty"`
	UserAgent map[string]interface{} `json:"userAgent,omitempty"`
}

// Response contains the items of response info log.
type Response struct {
	StatusCode int                    `json:"statusCode,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
}

// Host has the host information.
type Host struct {
	Hostname      string `json:"hostname,omitempty"`
	ForwardedHost string `json:"forwardedHost,omitempty"`
	IP            string `json:"ip,omitempty"`
}

// URL info
type URL struct {
	Path string `json:"path,omitempty"`
}

func removePort(host string) string {
	return strings.Split(host, ":")[0]
}

func getBodyLength(myw ReadableResponseWriter) int {
	if content := myw.Header().Get("Content-Length"); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return myw.Length()
}

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
		return http.HandlerFunc(HandleLog(LoggerHandler{
			Logger:         logger,
			ExcludedPrefix: excludedPrefix,
			OnNext: func(rrw *ReadableResponseWriter, httpRequest *http.Request) {
				next.ServeHTTP(rrw, httpRequest)
			},
		}))
	}
}
