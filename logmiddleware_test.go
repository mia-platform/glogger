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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gotest.tools/assert"
)

const hostname = "my-host.com"
const port = "3030"
const reqIDKey = "reqId"

var defaultRequestPath = fmt.Sprintf("http://%s:%s/my-req", hostname, port)

func testMockMiddlewareInvocation(next http.HandlerFunc, requestID string, logger *logrus.Logger, requestPath string) *test.Hook {
	if requestPath == "" {
		requestPath = defaultRequestPath
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Add("x-request-id", requestID)
	// create a null logger
	var hook *test.Hook
	if logger == nil {
		logger, hook = test.NewNullLogger()
	}
	if logger != nil {
		hook = test.NewLocal(logger)
	}
	handler := RequestMiddlewareLogger(logger, []string{"/-/"})
	// invoke the handler
	server := handler(next)
	// Create a response writer
	writer := httptest.NewRecorder()
	// Serve HTTP server
	server.ServeHTTP(writer, req)
	return hook
}

func assertJSON(t *testing.T, str string) error {
	var fields logrus.Fields

	err := json.Unmarshal([]byte(str), &fields)
	return err
}

type ExpectedLogFields struct {
	Level     logrus.Level
	RequestID string
	Message   string
}

type ExpectedIncomingLogFields struct {
	Method   string
	URL      string
	Hostname string
}

type ExpectedOutcomingLogFields struct {
	StatusCode int
}

func TestLogMiddleware(t *testing.T) {
	t.Run("create a middleware", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		testMockMiddlewareInvocation(handler, "", nil, "")

		assert.Assert(t, called, "handler is not called")
	})

	t.Run("log is a JSON also with trouble getting logger from context", func(t *testing.T) {
		var buffer bytes.Buffer
		logger, _ := InitHelper(InitOptions{})
		logger.Out = &buffer
		const logMessage = "New log message"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.WithValue(r.Context(), loggerKey{}, "notALogger")

			Get(r.Context()).Info(logMessage)
		})
		hook := testMockMiddlewareInvocation(handler, "", logger, "")

		assert.Equal(t, len(hook.AllEntries()), 3, "Number of logs is not 3")
		entry := hook.AllEntries()[1]
		assert.Equal(t, entry.Message, logMessage, "entry message is not the correct handler message")
		str := buffer.String()

		for i, value := range strings.Split(strings.TrimSpace(str), "\n") {
			err := assertJSON(t, value)
			assert.Equal(t, err, nil, "log %d is not a JSON", i)
		}
	})

	t.Run("middleware correctly passing configured logger with request id from request header", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMiddlewareInvocation(handler, requestID, nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:   http.MethodGet,
			URL:      defaultRequestPath,
			Hostname: hostname,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			StatusCode: statusCode,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")

		hook.Reset()
	})

	t.Run("middleware correctly passing configured logger with request id from request header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMiddlewareInvocation(handler, requestID, nil, "/-/healthz")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 0, "Unexpected entries length.")

		hook.Reset()
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMiddlewareInvocation(handler, "", nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:   logrus.InfoLevel,
			Message: "incoming request",
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:   http.MethodGet,
			URL:      defaultRequestPath,
			Hostname: hostname,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:   logrus.InfoLevel,
			Message: "request completed",
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			StatusCode: statusCode,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, fmt.Sprintf("Data reqId of request and response log must be the same. for log %d", i))

		hook.Reset()
	})
}

func logAssertions(t *testing.T, logEntry *logrus.Entry, expected ExpectedLogFields) string {
	assert.Equal(t, logEntry.Level, expected.Level, "Unexpected level of log for log in incoming request")
	assert.Equal(t, logEntry.Message, expected.Message, "Unexpected message of log for log in incoming request")
	requestID := logEntry.Data[reqIDKey]
	_, ok := requestID.(string)
	assert.Assert(t, ok, "Unexpected or empty requestID for log in incoming request")
	if expected.RequestID != "" {
		assert.Equal(t, requestID, expected.RequestID, "Unexpected requestID for log in incoming request")
	}
	return requestID.(string)
}

func incomingRequestAssertions(t *testing.T, incomingRequestLogEntry *logrus.Entry, expected ExpectedIncomingLogFields) {
	assert.Equal(t, incomingRequestLogEntry.Data["url"], expected.URL, "Unexpected http uri path for log in incoming request")
	assert.Equal(t, incomingRequestLogEntry.Data["method"], expected.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, incomingRequestLogEntry.Data["hostname"], expected.Hostname, "Unexpected hostname for log of request completed")
}

func outcomingRequestAssertions(t *testing.T, outcomingRequestLogEntry *logrus.Entry, expected ExpectedOutcomingLogFields) {
	assert.Equal(t, outcomingRequestLogEntry.Data["statusCode"], expected.StatusCode, "Unexpected status code for log of request completed")
	_, ok := outcomingRequestLogEntry.Data["responseTime"].(float64)
	assert.Assert(t, ok, "Invalid took duration for log of request completed")
}
