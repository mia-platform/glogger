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

package mux

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/glogger/v4"
	"github.com/mia-platform/glogger/v4/loggers/core"
	"github.com/mia-platform/glogger/v4/loggers/fake"
	"github.com/mia-platform/glogger/v4/middleware/utils"
	"github.com/stretchr/testify/require"
)

const hostname = "my-host.com"
const port = "3030"
const userAgent = "goHttp"
const bodyBytes = 0
const path = "/my-req"
const clientHost = "client-host"

const ip = "192.168.0.1"

var defaultRequestPath = fmt.Sprintf("http://%s:%s/my-req", hostname, port)

type ctxKey struct{}

func ctxMiddleware(ctx context.Context, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func testMockMuxMiddlewareInvocation(ctx context.Context, next http.HandlerFunc, requestID string, requestPath string) []fake.Record {
	if requestPath == "" {
		requestPath = defaultRequestPath
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Add("x-request-id", requestID)
	req.Header.Add("user-agent", userAgent)
	req.Header.Add("x-forwarded-for", ip)
	req.Header.Add("x-forwarded-host", clientHost)

	glog := fake.GetLogger()
	loggerMiddleware := RequestMiddlewareLogger(glog, []string{"/-/"})
	// invoke the handler
	server := loggerMiddleware(next)
	if ctx != nil {
		server = ctxMiddleware(ctx, server)
	}
	// Create a response writer
	writer := httptest.NewRecorder()
	// Serve HTTP server
	server.ServeHTTP(writer, req)
	return glog.OriginalLogger().AllRecords()
}

func TestMuxLogMiddleware(t *testing.T) {
	t.Run("create a middleware", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		testMockMuxMiddlewareInvocation(nil, handler, "", "")

		require.True(t, called, "handler is not called")
	})

	t.Run("middleware correctly passing request id from request header", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, "")
		require.Len(t, records, 2, "Unexpected entries length.")

		incomingRequest := records[0]
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method: http.MethodGet,
						UserAgent: utils.UserAgent{
							Original: userAgent,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
			},
			Message: utils.IncomingRequestMessage,
			Level:   "trace",
			Context: context.Background(),
		}, incomingRequest, "incoming request")

		outgoingRequest := records[1]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)

		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: bodyBytes,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("passing a content-length header by default", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Header().Set("content-length", "10")
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, "")
		require.Len(t, records, 2, "Unexpected entries length.")

		incomingRequest := records[0]
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
			},
			Message: utils.IncomingRequestMessage,
			Level:   "trace",
			Context: context.Background(),
		}, incomingRequest, "incoming request")

		outgoingRequest := records[1]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)

		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: 10,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("without content-length in the header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		contentToWrite := []byte("testing\n")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Write(contentToWrite)
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, "")
		require.Len(t, records, 2, "Unexpected entries length.")

		incomingRequest := records[0]
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
			},
			Message: utils.IncomingRequestMessage,
			Level:   "trace",
			Context: context.Background(),
		}, incomingRequest, "incoming request")

		outgoingRequest := records[1]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)

		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: len(contentToWrite),
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("request path without port", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		var requestPathWithoutPort = fmt.Sprintf("http://%s/my-req", hostname)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, requestPathWithoutPort)
		require.Len(t, records, 2, "Unexpected entries length.")

		outgoingRequest := records[1]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)
		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: bodyBytes,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("request path with query", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const pathWithQuery = "/my-req?foo=bar&some=other"
		var requestPathWithoutPort = fmt.Sprintf("http://%s%s", hostname, pathWithQuery)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})

		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, requestPathWithoutPort)
		require.Len(t, records, 2, "Unexpected entries length.")

		outgoingRequest := records[1]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)
		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: bodyBytes,
						},
					},
				},
				"url": utils.URL{Path: pathWithQuery},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("exclude prefix endpoints", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, requestID, "/-/healthz")
		require.Len(t, records, 0, "Unexpected entries length.")
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		records := testMockMuxMiddlewareInvocation(nil, handler, "", "")
		require.Len(t, records, 2, "Unexpected entries length.")

		incomingRequestReqId := records[0].Fields["reqId"].(string)
		require.NotEmpty(t, incomingRequestReqId)

		requestCompletedReqId := records[1].Fields["reqId"].(string)
		require.NotEmpty(t, requestCompletedReqId)

		require.Equal(t, incomingRequestReqId, requestCompletedReqId)
	})

	t.Run("middleware correctly pass context to logger", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"

		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKey{}, "ok")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := glogger.GetOrDie[core.Logger[*fake.Entry]](r.Context())
			logger.WithFields(map[string]any{"foo": "bar"}).Info("ok")

			w.WriteHeader(statusCode)
		})
		records := testMockMuxMiddlewareInvocation(ctx, handler, requestID, "")
		require.Len(t, records, 3, "Unexpected entries length.")

		incomingRequest := records[0]
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method: http.MethodGet,
						UserAgent: utils.UserAgent{
							Original: userAgent,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
			},
			Message: utils.IncomingRequestMessage,
			Level:   "trace",
			Context: ctx,
		}, incomingRequest, "incoming request")

		handlerLog := records[1]
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"foo":   "bar",
			},
			Message: "ok",
			Level:   "info",
			Context: ctx,
		}, handlerLog, "handler log")

		outgoingRequest := records[2]
		require.InDelta(t, 100, outgoingRequest.Fields["responseTime"], 100)

		outgoingRequest.Fields["responseTime"] = 0
		require.Equal(t, fake.Record{
			Fields: map[string]any{
				"reqId": requestID,
				"http": utils.HTTP{
					Request: &utils.Request{
						Method:    http.MethodGet,
						UserAgent: utils.UserAgent{Original: userAgent},
					},
					Response: &utils.Response{
						StatusCode: statusCode,
						Body: utils.ResponseBody{
							Bytes: bodyBytes,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      hostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: ctx,
		}, outgoingRequest)
	})
}
