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

package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mia-platform/glogger/v3"
	"github.com/mia-platform/glogger/v3/loggers/core"
)

const (
	forwardedHostHeaderKey = "x-forwarded-host"
	forwardedForHeaderKey  = "x-forwarded-for"
	requestIDHeaderName    = "x-request-id"

	IncomingRequestMessage  = "incoming request"
	RequestCompletedMessage = "request completed"
)

// HTTP is the struct of the log formatter.
type HTTP struct {
	Request  *Request  `json:"request,omitempty"`
	Response *Response `json:"response,omitempty"`
}

type UserAgent struct {
	Original string `json:"original,omitempty"`
}

// Request contains the items of request info log.
type Request struct {
	Method    string    `json:"method,omitempty"`
	UserAgent UserAgent `json:"userAgent,omitempty"`
}

type ResponseBody struct {
	Bytes int `json:"bytes,omitempty"`
}

// Response contains the items of response info log.
type Response struct {
	StatusCode int          `json:"statusCode,omitempty"`
	Body       ResponseBody `json:"body,omitempty"`
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

func GetReqID(ctx glogger.LoggingContext) string {
	if requestID := ctx.Request().GetHeader(requestIDHeaderName); requestID != "" {
		return requestID
	}
	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
	requestID, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Errorf("error generating request id: %s", err))
	}
	return requestID.String()
}

func LogIncomingRequest[T any](ctx glogger.LoggingContext, logger core.Logger[T]) {
	logger.
		WithFields(map[string]any{
			"http": HTTP{
				Request: &Request{
					Method: ctx.Request().Method(),
					UserAgent: UserAgent{
						Original: ctx.Request().GetHeader("user-agent"),
					},
				},
			},
			"url": URL{Path: ctx.Request().URI()},
			"host": Host{
				ForwardedHost: ctx.Request().GetHeader(forwardedHostHeaderKey),
				Hostname:      removePort(ctx.Request().Host()),
				IP:            ctx.Request().GetHeader(forwardedForHeaderKey),
			},
		}).
		Trace(IncomingRequestMessage)
}

func LogRequestCompleted[T any](ctx glogger.LoggingContext, logger core.Logger[T], startTime time.Time) {
	logger.
		WithFields(map[string]any{
			"http": HTTP{
				Request: &Request{
					Method: ctx.Request().Method(),
					UserAgent: UserAgent{
						Original: ctx.Request().GetHeader("user-agent"),
					},
				},
				Response: &Response{
					StatusCode: ctx.Response().StatusCode(),
					Body: ResponseBody{
						Bytes: ctx.Response().BodySize(),
					},
				},
			},
			"url": URL{Path: ctx.Request().URI()},
			"host": Host{
				ForwardedHost: ctx.Request().GetHeader(forwardedHostHeaderKey),
				Hostname:      removePort(ctx.Request().Host()),
				IP:            ctx.Request().GetHeader(forwardedForHeaderKey),
			},
			"responseTime": float64(time.Since(startTime).Milliseconds()),
		}).
		Info(RequestCompletedMessage)
}
