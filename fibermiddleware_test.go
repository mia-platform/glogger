package glogger

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/valyala/fasthttp"
	"gotest.tools/assert"
)

func testMockFiberMiddlewareInvocation(handler fiber.Handler, requestID string, logger *logrus.Logger, requestPath string) (*test.Hook, error) {
	if requestPath == "" {
		requestPath = "/my-req"
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Add("x-request-id", requestID)
	req.Header.Add("user-agent", userAgent)
	req.Header.Add("x-forwarded-for", ip)
	req.Header.Add("x-forwarded-host", clientHost)
	ip = removePort(req.RemoteAddr)

	// create a null logger
	var hook *test.Hook
	if logger == nil {
		logger, hook = test.NewNullLogger()
		logger.SetLevel(logrus.TraceLevel)
	}
	if logger != nil {
		hook = test.NewLocal(logger)
	}

	// invoke the middleware
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)
	app.Use(RequestFiberMiddlewareLogger(logger, []string{"/-/"}))
	app.Get(requestPath, handler)

	_, err := app.Test(req)

	return hook, err
}

func TestFiberLogMiddleware(t *testing.T) {
	hostname := "example.com"
	t.Run("create a middleware", func(t *testing.T) {
		called := false
		testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			called = true
			return nil
		}, "", nil, "")

		assert.Assert(t, called, "handler is not called")
	})

	t.Run("log is a JSON also with trouble getting logger from context", func(t *testing.T) {
		var buffer bytes.Buffer
		logger, _ := InitHelper(InitOptions{Level: "trace"})
		logger.Out = &buffer
		const logMessage = "New log message"
		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			ctx := context.WithValue(c.UserContext(), loggerKey{}, "notALogger")
			Get(ctx).Info(logMessage)
			return nil
		}, "", logger, "")

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

		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, nil, "")

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

		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Set("content-length", "10")
			return nil
		}, requestID, nil, "")

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

		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Write(contentToWrite)
			return nil
		}, requestID, nil, "")

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

		logger, _ := test.NewNullLogger()
		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, "")

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

		logger, _ := test.NewNullLogger()
		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, requestPathWithoutPort)

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

		logger, _ := InitHelper(InitOptions{
			DisableHTMLEscape: true,
		})
		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, pathWithQuery)

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

		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, nil, "/-/healthz")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 0, "Unexpected entries length.")

		hook.Reset()
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400

		hook, _ := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, "", nil, "")

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
