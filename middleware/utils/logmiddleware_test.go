/*
 * Copyright 2023 Mia srl
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

package utils

import (
	"context"
	"net/http"
	"testing"
	"time"

	fakeLogger "github.com/mia-platform/glogger/v3/loggers/fake"
	"github.com/mia-platform/glogger/v3/middleware/utils/internal/fake"
	"github.com/stretchr/testify/require"
)

func TestGetReqId(t *testing.T) {
	t.Run("get req id from header", func(t *testing.T) {
		reqId := "my-req-id"
		logCtx := fake.NewContext(context.Background(), fake.Request{
			Headers: map[string]string{
				requestIDHeaderName: reqId,
			},
		}, fake.Response{})

		require.Equal(t, reqId, GetReqID(logCtx))
	})

	t.Run("generate req id if not taken from header", func(t *testing.T) {
		logCtx := fake.NewContext(context.Background(), fake.Request{}, fake.Response{})

		require.NotEmpty(t, GetReqID(logCtx))
	})
}

func TestLogIncomingRequest(t *testing.T) {
	t.Run("log incoming request correctly", func(t *testing.T) {
		ctx := fake.NewContext(context.Background(), fake.Request{
			Headers: map[string]string{
				"user-agent":           "my-agent",
				forwardedHostHeaderKey: "my-host",
				forwardedForHeaderKey:  "127.0.0.1",
			},
		}, fake.Response{})
		logger := fakeLogger.GetLogger()

		LogIncomingRequest(ctx, logger)

		records := logger.GetOriginalLogger().AllRecords()
		require.Equal(t, fakeLogger.Record{
			Level:   "trace",
			Message: IncomingRequestMessage,
			Fields: map[string]any{
				"http": HTTP{
					Request: &Request{
						Method:    "GET",
						UserAgent: map[string]interface{}{"original": "my-agent"},
					},
				},
				"url": URL{Path: "/custom-uri"},
				"host": Host{
					ForwardedHost: "my-host",
					Hostname:      "echo-service",
					IP:            "127.0.0.1",
				},
			},
		}, records[0])
	})
}

func TestLogRequestCompleted(t *testing.T) {
	t.Run("request completed log correctly", func(t *testing.T) {
		ctx := fake.NewContext(context.Background(), fake.Request{
			Headers: map[string]string{
				"user-agent":           "my-agent",
				forwardedHostHeaderKey: "my-host",
				forwardedForHeaderKey:  "127.0.0.1",
			},
		}, fake.Response{
			StatusCode: http.StatusOK,
			BodySize:   12,
		})
		logger := fakeLogger.GetLogger()

		startTime := time.Now()

		LogRequestCompleted(ctx, logger, startTime)

		records := logger.GetOriginalLogger().AllRecords()
		require.Equal(t, fakeLogger.Record{
			Level:   "info",
			Message: RequestCompletedMessage,
			Fields: map[string]any{
				"http": HTTP{
					Request: &Request{
						Method:    "GET",
						UserAgent: map[string]interface{}{"original": "my-agent"},
					},
					Response: &Response{
						StatusCode: http.StatusOK,
						Body: map[string]interface{}{
							"bytes": 12,
						},
					},
				},
				"url": URL{Path: "/custom-uri"},
				"host": Host{
					ForwardedHost: "my-host",
					Hostname:      "echo-service",
					IP:            "127.0.0.1",
				},
				"responseTime": records[0].Fields["responseTime"],
			},
		}, records[0])
		require.InDelta(t, records[0].Fields["responseTime"], 0, 1000)
	})
}
