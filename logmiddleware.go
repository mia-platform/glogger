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
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func getReqID(logger *logrus.Logger, getHeader func(string) string) string {
	if requestID := getHeader("X-Request-Id"); requestID != "" {
		return requestID
	}
	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
	requestID, err := uuid.NewRandom()
	if err != nil {
		logger.WithError(err).Fatal("error generating request id")
	}
	return requestID.String()
}

func logBeforeHandler(ctx LoggingContext) {
	Get(ctx.Context()).WithFields(logrus.Fields{
		"http": HTTP{
			Request: &Request{
				Method:    ctx.Request().Method(),
				UserAgent: map[string]interface{}{"original": ctx.Request().GetHeader("user-agent")},
			},
		},
		"url": URL{Path: ctx.Request().URI()},
		"host": Host{
			ForwardedHost: ctx.Request().GetHeader(forwardedHostHeaderKey),
			Hostname:      removePort(ctx.Request().Host()),
			IP:            ctx.Request().GetHeader(forwardedForHeaderKey),
		},
	}).Trace("incoming request")
}

func logAfterHandler(ctx LoggingContext, startTime time.Time) {
	Get(ctx.Context()).WithFields(logrus.Fields{
		"http": HTTP{
			Request: &Request{
				Method:    ctx.Request().Method(),
				UserAgent: map[string]interface{}{"original": ctx.Request().GetHeader("user-agent")},
			},
			Response: &Response{
				StatusCode: ctx.Response().StatusCode(),
				Body: map[string]interface{}{
					"bytes": ctx.Response().BodySize(),
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
	}).Info("request completed")
}

func RequestFiberMiddlewareLogger(logger *logrus.Logger, excludedPrefix []string) func(*fiber.Ctx) error {
	return func(fiberCtx *fiber.Ctx) error {
		requestURI := fiberCtx.Request().URI().String()

		for _, prefix := range excludedPrefix {
			if strings.HasPrefix(requestURI, prefix) {
				return fiberCtx.Next()
			}
		}

		start := time.Now()

		requestID := getReqID(logger, func(name string) string { return fiberCtx.Get(name, "") })
		ctx := WithLogger(fiberCtx.UserContext(), logrus.NewEntry(logger).WithFields(logrus.Fields{
			"reqId": requestID,
		}))
		fiberCtx.SetUserContext(ctx)

		Get(ctx).WithFields(logrus.Fields{
			"http": HTTP{
				Request: &Request{
					Method:    fiberCtx.Method(),
					UserAgent: map[string]interface{}{"original": fiberCtx.Get("user-agent")},
				},
			},
			"url": URL{Path: requestURI},
			"host": Host{
				ForwardedHost: fiberCtx.Get(forwardedHostHeaderKey),
				Hostname:      removePort(string(fiberCtx.Request().Host())),
				IP:            fiberCtx.Get(forwardedForHeaderKey),
			},
		}).Trace("incoming request")

		err := fiberCtx.Next()

		Get(ctx).WithFields(logrus.Fields{
			"http": HTTP{
				Request: &Request{
					Method:    fiberCtx.Method(),
					UserAgent: map[string]interface{}{"original": fiberCtx.Get("user-agent")},
				},
				Response: &Response{
					StatusCode: fiberCtx.Response().StatusCode(),
					Body: map[string]interface{}{
						"bytes": len(fiberCtx.Response().Body()),
					},
				},
			},
			"url": URL{Path: requestURI},
			"host": Host{
				ForwardedHost: fiberCtx.Get(forwardedHostHeaderKey),
				Hostname:      removePort(string(fiberCtx.Request().Host())),
				IP:            fiberCtx.Get(forwardedForHeaderKey),
			},
			"responseTime": float64(time.Since(start).Milliseconds()),
		}).Info("request completed")

		return err
	}
}
