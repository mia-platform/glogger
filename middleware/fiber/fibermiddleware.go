package fiber

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/glogger/v3"
	"github.com/mia-platform/glogger/v3/loggers/core"
	"github.com/mia-platform/glogger/v3/middleware/utils"
)

type fiberLoggingContext struct {
	c          *fiber.Ctx
	handlerErr error
}

func (flc *fiberLoggingContext) Request() glogger.RequestLoggingContext {
	return flc
}

func (flc *fiberLoggingContext) Response() glogger.ResponseLoggingContext {
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

func (flc fiberLoggingContext) getFiberError() *fiber.Error {
	if fiberErr, ok := flc.handlerErr.(*fiber.Error); flc.handlerErr != nil && ok {
		return fiberErr
	}
	return nil
}

func (flc *fiberLoggingContext) setError(err error) {
	flc.handlerErr = err
}

func (flc *fiberLoggingContext) BodySize() int {
	if fiberErr := flc.getFiberError(); fiberErr != nil {
		return len(fiberErr.Error())
	}

	if content := flc.c.GetRespHeader("Content-Length"); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return len(flc.c.Response().Body())
}

func (flc *fiberLoggingContext) StatusCode() int {
	if fiberErr := flc.getFiberError(); fiberErr != nil {
		return fiberErr.Code
	}

	return flc.c.Response().StatusCode()
}

// RequestMiddlewareLogger is a fiber middleware to log all requests
// It logs the incoming request and when request is completed, adding latency of the request
func RequestMiddlewareLogger[Logger any](logger core.Logger[Logger], excludedPrefix []string) func(*fiber.Ctx) error {
	return func(fiberCtx *fiber.Ctx) error {
		fiberLoggingContext := &fiberLoggingContext{c: fiberCtx}

		for _, prefix := range excludedPrefix {
			if strings.HasPrefix(fiberLoggingContext.Request().URI(), prefix) {
				return fiberCtx.Next()
			}
		}

		start := time.Now()

		requestID := utils.GetReqID(fiberLoggingContext)
		loggerWithReqId := logger.WithFields(map[string]any{
			"reqId": requestID,
		})
		ctx := glogger.WithLogger(fiberCtx.UserContext(), loggerWithReqId.OriginalLogger())
		fiberCtx.SetUserContext(ctx)

		utils.LogIncomingRequest[Logger](fiberLoggingContext, loggerWithReqId)
		err := fiberCtx.Next()
		fiberLoggingContext.setError(err)

		utils.LogRequestCompleted[Logger](fiberLoggingContext, loggerWithReqId, start)

		return err
	}
}
