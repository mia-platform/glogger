package integrationtest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	glogrus "github.com/mia-platform/glogger/v3/loggers/logrus"
	"github.com/mia-platform/glogger/v3/loggers/logrus/testhttplog"
	gloggerfiber "github.com/mia-platform/glogger/v3/middleware/fiber"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestLogrusFiber(t *testing.T) {
	mockHostname := "example.com"

	t.Run("middleware correctly log", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"

		hook := testMockFiberMiddlewareInvocation(func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, nil, mockHostname, "")

		entries := hook.AllEntries()
		require.Len(t, entries, 2, "Unexpected entries length.")

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

		require.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")

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
		require.Len(t, entries, 1, "Unexpected entries length.")

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

}

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

	// invoke the middleware
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	glog := glogrus.GetLogger(logrus.NewEntry(logger))
	app.Use(gloggerfiber.RequestFiberMiddlewareLogger(glog, []string{"/-/"}))

	requestPathWithoutQuery := strings.Split(requestPath, "?")[0]
	app.Get(requestPathWithoutQuery, handler)

	app.Test(req)

	return hook
}
