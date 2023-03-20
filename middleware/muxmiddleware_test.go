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

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mia-platform/glogger/v3"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gotest.tools/assert"
)

const hostname = "my-host.com"
const port = "3030"
const reqIDKey = "reqId"
const userAgent = "goHttp"
const bodyBytes = 0
const path = "/my-req"
const clientHost = "client-host"

const ip = "192.168.0.1"

var defaultRequestPath = fmt.Sprintf("http://%s:%s/my-req", hostname, port)

func testMockMuxMiddlewareInvocation(next http.HandlerFunc, requestID string, logger *logrus.Logger, requestPath string) *test.Hook {
	if requestPath == "" {
		requestPath = defaultRequestPath
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Add("x-request-id", requestID)
	req.Header.Add("user-agent", userAgent)
	req.Header.Add("x-forwarded-for", ip)
	req.Header.Add("x-forwarded-host", clientHost)

	// create a null logger
	var hook *test.Hook
	if logger == nil {
		logger, hook = test.NewNullLogger()
		logger.SetLevel(logrus.TraceLevel)
	}
	if logger != nil {
		hook = test.NewLocal(logger)
	}
	handler := RequestGorillaMuxMiddlewareLogger(logger, []string{"/-/"})
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
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
}

type ExpectedOutcomingLogFields struct {
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
	Bytes         int
	StatusCode    int
}

func TestMuxLogMiddleware(t *testing.T) {
	t.Run("create a middleware", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		testMockMuxMiddlewareInvocation(handler, "", nil, "")

		assert.Assert(t, called, "handler is not called")
	})

	t.Run("log is a JSON also with trouble getting logger from context", func(t *testing.T) {
		var buffer bytes.Buffer
		logger, _ := glogger.InitHelper(glogger.InitOptions{Level: "trace"})
		logger.Out = &buffer
		const logMessage = "New log message"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			glogger.Get(context.Background()).Info(logMessage)
		})
		hook := testMockMuxMiddlewareInvocation(handler, "", logger, "")

		assert.Equal(t, len(hook.AllEntries()), 2, "Number of logs is not 2")
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
		hook := testMockMuxMiddlewareInvocation(handler, requestID, nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")

		hook.Reset()
	})

	t.Run("passing a content-length header by default", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Header().Set("content-length", "10")
		})
		hook := testMockMuxMiddlewareInvocation(handler, requestID, nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         10,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")

		hook.Reset()
	})

	t.Run("without content-length in the header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		contentToWrite := []byte("testing\n")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Write(contentToWrite)
		})
		hook := testMockMuxMiddlewareInvocation(handler, requestID, nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         len(contentToWrite),
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")

		hook.Reset()
	})

	t.Run("using info level returning only outcomingRequest", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger, _ := test.NewNullLogger()
		hook := testMockMuxMiddlewareInvocation(handler, requestID, logger, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		hook.Reset()
	})

	t.Run("test getHostname with request path without port", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		var requestPathWithoutPort = fmt.Sprintf("http://%s/my-req", hostname)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger, _ := test.NewNullLogger()
		hook := testMockMuxMiddlewareInvocation(handler, requestID, logger, requestPathWithoutPort)

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		hook.Reset()
	})

	t.Run("test getHostname with request path with query", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const pathWithQuery = "/my-req?foo=bar&some=other"
		var requestPathWithoutPort = fmt.Sprintf("http://%s%s", hostname, pathWithQuery)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger, _ := glogger.InitHelper(glogger.InitOptions{
			DisableHTMLEscape: true,
		})
		hook := testMockMuxMiddlewareInvocation(handler, requestID, logger, requestPathWithoutPort)

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")
		byteEntry, err := entries[0].Bytes()
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(string(byteEntry), `"url":{"path":"/my-req?foo=bar&some=other"}`))
		hook.Reset()
	})

	t.Run("middleware correctly passing configured logger with request id from request header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMuxMiddlewareInvocation(handler, requestID, nil, "/-/healthz")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 0, "Unexpected entries length.")

		hook.Reset()
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMuxMiddlewareInvocation(handler, "", nil, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "incoming request",
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:   logrus.InfoLevel,
			Message: "request completed",
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, fmt.Sprintf("Data reqId of request and response log must be the same. for log %d", i))

		hook.Reset()
	})
}

func logAssertions(t *testing.T, logEntry *logrus.Entry, expected ExpectedLogFields) string {
	t.Helper()
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
	http := incomingRequestLogEntry.Data["http"].(HTTP)
	assert.Equal(t, http.Request.Method, expected.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original, "Unexpected original userAgent for log of request completed")

	url := incomingRequestLogEntry.Data["url"].(URL)
	assert.Equal(t, url.Path, expected.Path, "Unexpected http uri path for log in incoming request")

	host := incomingRequestLogEntry.Data["host"].(Host)
	assert.Equal(t, host.Hostname, expected.Hostname, "Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP, "Unexpected ip for log of request completed")
}

func outcomingRequestAssertions(t *testing.T, outcomingRequestLogEntry *logrus.Entry, expected ExpectedOutcomingLogFields) {
	http := outcomingRequestLogEntry.Data["http"].(HTTP)
	assert.Equal(t, http.Request.Method, expected.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original, "Unexpected original userAgent for log of request completed")
	assert.Equal(t, http.Response.StatusCode, expected.StatusCode, "Unexpected status code for log of request completed")
	assert.Equal(t, http.Response.Body["bytes"], expected.Bytes, "Unexpected status code for log of request completed")

	url := outcomingRequestLogEntry.Data["url"].(URL)
	assert.Equal(t, url.Path, expected.Path, "Unexpected http uri path for log in incoming request")

	host := outcomingRequestLogEntry.Data["host"].(Host)
	assert.Equal(t, host.Hostname, expected.Hostname, "Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP, "Unexpected ip for log of request completed")

	_, ok := outcomingRequestLogEntry.Data["responseTime"].(float64)
	assert.Assert(t, ok, "Invalid took duration for log of request completed")
}
