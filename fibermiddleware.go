package glogger

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type fiberLoggingContext struct {
	c *fiber.Ctx
}

func (flc *fiberLoggingContext) Context() context.Context {
	return flc.c.UserContext()
}

func (flc *fiberLoggingContext) Request() RequestLoggingContext {
	return flc
}

func (flc *fiberLoggingContext) Response() ResponseLoggingContext {
	return flc
}

func (flc *fiberLoggingContext) GetHeader(key string) string {
	return flc.c.Get(key, "")
}

func (flc *fiberLoggingContext) URI() string {
	return string(flc.c.Request().URI().PathOriginal())
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

		fiberLoggingContext := &fiberLoggingContext{fiberCtx}

		logBeforeHandler(fiberLoggingContext)
		err := fiberCtx.Next()
		logAfterHandler(fiberLoggingContext, start)

		return err
	}
}
