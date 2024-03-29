package fiber

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/glogger/v4"
	"github.com/mia-platform/glogger/v4/loggers/core"
	"github.com/mia-platform/glogger/v4/loggers/fake"
	"github.com/mia-platform/glogger/v4/middleware/utils"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

const userAgent = "goHttp"
const bodyBytes = 0
const path = "/my-req"
const clientHost = "client-host"

const ip = "192.168.0.1"

type ctxKey struct{}

func ctxMiddleware(ctx context.Context) func(*fiber.Ctx) error {
	return func(fiberCtx *fiber.Ctx) error {
		fiberCtx.SetUserContext(ctx)
		return fiberCtx.Next()
	}
}

func testMockFiberMiddlewareInvocation(ctx context.Context, handler fiber.Handler, requestID string, hostname, requestPath string) []fake.Record {
	if requestPath == "" {
		requestPath = path
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

	glog := fake.GetLogger()

	// invoke the middleware
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	if ctx != nil {
		app.Use(ctxMiddleware(ctx))
	}
	app.Use(RequestMiddlewareLogger(glog, []string{"/-/"}))

	requestPathWithoutQuery := strings.Split(requestPath, "?")[0]
	app.Get(requestPathWithoutQuery, handler)

	app.Test(req)

	return glog.OriginalLogger().AllRecords()
}

func TestFiberLogMiddleware(t *testing.T) {
	mockHostname := "example.com"

	t.Run("create a middleware", func(t *testing.T) {
		called := false
		testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			called = true
			return nil
		}, "", mockHostname, "")

		require.True(t, called, "handler is not called")
	})

	t.Run("request id from request header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const reqPath = "/my-req"

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, mockHostname, reqPath)
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
				"url": utils.URL{Path: reqPath},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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
				"url": utils.URL{Path: reqPath},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, mockHostname, pathWithQuery)
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
				"url": utils.URL{Path: pathWithQuery},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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
				"url": utils.URL{Path: pathWithQuery},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)
	})

	t.Run("request on non-existing route should cause a 404 log", func(t *testing.T) {
		requestPath := "/non-existing"
		requestID := "someId"

		req := httptest.NewRequest(http.MethodGet, requestPath, nil)
		req.Header.Add("x-request-id", requestID)
		req.Header.Add("user-agent", userAgent)
		req.Header.Add("x-forwarded-for", ip)
		req.Header.Add("x-forwarded-host", clientHost)

		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)

		glog := fake.GetLogger()
		app.Use(RequestMiddlewareLogger(glog, []string{"/-/"}))
		app.Test(req)

		records := glog.OriginalLogger().AllRecords()
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
						StatusCode: 404,
						Body: utils.ResponseBody{
							Bytes: len("Cannot GET /non-existing"),
						},
					},
				},
				"url": utils.URL{Path: requestPath},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Set("content-length", "10")
			return nil
		}, requestID, mockHostname, "")
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
							Bytes: 10,
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			c.Write(contentToWrite)
			return nil
		}, requestID, mockHostname, "")
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
							Bytes: len(contentToWrite),
						},
					},
				},
				"url": utils.URL{Path: path},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
					IP:            ip,
				},
				"responseTime": 0,
			},
			Message: utils.RequestCompletedMessage,
			Level:   "info",
			Context: context.Background(),
		}, outgoingRequest)

	})

	t.Run("do not log skipped paths", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, requestID, mockHostname, "/-/healthz")

		require.Len(t, records, 0, "Unexpected entries length.")
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400

		records := testMockFiberMiddlewareInvocation(nil, func(c *fiber.Ctx) error {
			c.Status(statusCode)
			return nil
		}, "", mockHostname, "")
		require.Len(t, records, 2, "Unexpected entries length.")

		incomingRequestReqId := records[0].Fields["reqId"].(string)
		require.NotEmpty(t, incomingRequestReqId)

		requestCompletedReqId := records[1].Fields["reqId"].(string)
		require.NotEmpty(t, requestCompletedReqId)

		require.Equal(t, incomingRequestReqId, requestCompletedReqId)
	})

	t.Run("passing user context in logger", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const reqPath = "/my-req"
		ctx := context.WithValue(context.Background(), ctxKey{}, "ok")

		records := testMockFiberMiddlewareInvocation(ctx, func(c *fiber.Ctx) error {
			c.Status(statusCode)

			logger := glogger.GetOrDie[core.Logger[*fake.Entry]](c.UserContext())
			logger.WithFields(map[string]any{"foo": "bar"}).Info("ok")

			return nil
		}, requestID, mockHostname, reqPath)
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
				"url": utils.URL{Path: reqPath},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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
				"url": utils.URL{Path: reqPath},
				"host": utils.Host{
					ForwardedHost: clientHost,
					Hostname:      mockHostname,
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
