package middleware

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/glogger/v3"
	"github.com/sirupsen/logrus"
)

type fiberLoggingContext struct {
	c *fiber.Ctx
}

func (flc *fiberLoggingContext) Context() context.Context {
	return flc.c.UserContext()
}

func (flc *fiberLoggingContext) Request() requestLoggingContext {
	return flc
}

func (flc *fiberLoggingContext) Response() responseLoggingContext {
	return flc
}

func (flc *fiberLoggingContext) GetHeader(key string) string {
	return flc.c.Get(key, "")
}

func (flc *fiberLoggingContext) URI() string {
	return string(flc.c.Request().URI().RequestURI())
}

func (flc *fiberLoggingContext) Host() string {
	return string(flc.c.Request().Host())
}

func (flc *fiberLoggingContext) Method() string {
	return flc.c.Method()
}

func (flc *fiberLoggingContext) BodySize() int {
	if content := flc.c.GetRespHeader("Content-Length"); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return len(flc.c.Response().Body())
}

func (flc *fiberLoggingContext) StatusCode() int {
	return flc.c.Response().StatusCode()
}

// RequestFiberMiddlewareLogger is a fiber middleware to log all requests with logrus
// It logs the incoming request and when request is completed, adding latency of the request
func RequestFiberMiddlewareLogger(logger *logrus.Logger, excludedPrefix []string) func(*fiber.Ctx) error {
	return func(fiberCtx *fiber.Ctx) error {
		fiberLoggingContext := &fiberLoggingContext{fiberCtx}

		for _, prefix := range excludedPrefix {
			if strings.HasPrefix(fiberLoggingContext.Request().URI(), prefix) {
				return fiberCtx.Next()
			}
		}

		start := time.Now()

		requestID := getReqID(logger, fiberLoggingContext)
		ctx := glogger.WithLogger(fiberCtx.UserContext(), logrus.NewEntry(logger).WithFields(logrus.Fields{
			"reqId": requestID,
		}))
		fiberCtx.SetUserContext(ctx)

		logBeforeHandler(fiberLoggingContext)
		err := fiberCtx.Next()
		logAfterHandler(fiberLoggingContext, start, err)

		return err
	}
}
