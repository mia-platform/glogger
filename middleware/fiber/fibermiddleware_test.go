package fiber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v3/loggers/logrus"
	"github.com/mia-platform/glogger/v3/loggers/logrus/testhttplog"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/valyala/fasthttp"
	"gotest.tools/assert"
)

const reqIDKey = "reqId"
const userAgent = "goHttp"
const bodyBytes = 0
const path = "/my-req"
const clientHost = "client-host"

const ip = "192.168.0.1"

func testMockFiberMiddlewareInvocation(handler fiber.Handler, requestID string, logger *logrus.Logger, hostname, requestPath string) *test.Hook {
	if requestPath == "" {
		requestPath = "/my-req"
	}
	// create a request
	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://%s%s", hostname, requestPath),
		nil,
	)
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

	glog := glogrus.GetLogger(logrus.NewEntry(logger))

	// invoke the middleware
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	app.Use(RequestFiberMiddlewareLogger(glog, []string{"/-/"}))

	requestPathWithoutQuery := strings.Split(requestPath, "?")[0]
	app.Get(requestPathWithoutQuery, handler)

	app.Test(req)

	return hook
}

func TestFiberLogMiddleware(t *testing.T) {
	mockHostname := "example.com"

	t.Run("test getHostname with request path without port", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const reqPath = "/my-req"

		logger, _ := test.NewNullLogger()
		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, mockHostname, reqPath)

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
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

		logger, _ := glogrus.InitHelper(glogrus.InitOptions{
			DisableHTMLEscape: true,
		})
		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, mockHostname, pathWithQuery)

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")
		byteEntry, err := entries[0].Bytes()
		assert.NilError(t, err)
		assert.Check(t, strings.Contains(string(byteEntry), `"url":{"path":"/my-req?foo=bar&some=other"}`))

		hook.Reset()
	})

	t.Run("request on non-existing route should cause a 404 log", func(t *testing.T) {
		requestPath := "/non-existing"
		requestID := "someId"
		logger, hook := test.NewNullLogger()
		logger.SetLevel(logrus.TraceLevel)
		hook.Reset()

		req := httptest.NewRequest(http.MethodGet, requestPath, nil)
		req.Header.Add("x-request-id", requestID)
		req.Header.Add("user-agent", userAgent)
		req.Header.Add("x-forwarded-for", ip)
		req.Header.Add("x-forwarded-host", clientHost)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		glog := glogrus.GetLogger(logrus.NewEntry(logger))
		app.Use(RequestFiberMiddlewareLogger(glog, []string{"/-/"}))
		app.Test(req)

		logEntries := hook.AllEntries()
		lastEntry := logEntries[len(logEntries)-1]
		testhttplog.OutgoingRequestAssertions(t, lastEntry, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          requestPath,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    http.StatusNotFound,
			Bytes:         len("Cannot GET /non-existing"),
		})
	})

	t.Run("create a middleware", func(t *testing.T) {
		called := false
		testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			called = true
			return nil
		}, "", nil, mockHostname, "")

		assert.Assert(t, called, "handler is not called")
	})

	t.Run("log is a JSON also with trouble getting logger from context", func(t *testing.T) {
		var buffer bytes.Buffer
		logger, _ := glogrus.InitHelper(glogrus.InitOptions{Level: "trace"})
		logger.Out = &buffer
		const logMessage = "New log message"
		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			glogrus.GetFromContext(context.Background()).Info(logMessage)
			return nil
		}, "", logger, mockHostname, "")

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

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, nil, mockHostname, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := testhttplog.LogAssertions(t, incomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		testhttplog.IncomingRequestAssertions(t, incomingRequest, testhttplog.ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
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

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Set("content-length", "10")
			return nil
		}, requestID, nil, mockHostname, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := testhttplog.LogAssertions(t, incomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		testhttplog.IncomingRequestAssertions(t, incomingRequest, testhttplog.ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
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

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Write(contentToWrite)
			return nil
		}, requestID, nil, mockHostname, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := testhttplog.LogAssertions(t, incomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.TraceLevel,
			Message:   "incoming request",
			RequestID: requestID,
		})
		testhttplog.IncomingRequestAssertions(t, incomingRequest, testhttplog.ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
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
		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, logger, mockHostname, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:     logrus.InfoLevel,
			Message:   "request completed",
			RequestID: requestID,
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		hook.Reset()
	})

	t.Run("do not log skipped paths", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, nil, mockHostname, "/-/healthz")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 0, "Unexpected entries length.")

		hook.Reset()
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, "", nil, mockHostname, "")

		entries := hook.AllEntries()
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := testhttplog.LogAssertions(t, incomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "incoming request",
		})
		assert.Assert(t, incomingRequestID != "")
		testhttplog.IncomingRequestAssertions(t, incomingRequest, testhttplog.ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := testhttplog.LogAssertions(t, outcomingRequest, reqIDKey, testhttplog.ExpectedLogFields{
			Level:   logrus.InfoLevel,
			Message: "request completed",
		})
		testhttplog.OutgoingRequestAssertions(t, outcomingRequest, testhttplog.ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      mockHostname,
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

func assertJSON(t *testing.T, str string) error {
	var fields logrus.Fields

	err := json.Unmarshal([]byte(str), &fields)
	return err
}
