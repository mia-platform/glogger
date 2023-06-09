package integrationtest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	glogrus "github.com/mia-platform/glogger/v3/loggers/logrus"
	"github.com/mia-platform/glogger/v3/loggers/logrus/testhttplog"
	"github.com/mia-platform/glogger/v3/middleware/mux"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestLogrusMux(t *testing.T) {
	t.Run("middleware correctly log", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		hook := testMockMuxMiddlewareInvocation(handler, requestID, nil, "")

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
			Hostname:      hostname,
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
			Hostname:      hostname,
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
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger, _ := test.NewNullLogger()
		hook := testMockMuxMiddlewareInvocation(handler, requestID, logger, "")

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
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		hook.Reset()
	})
}

func testMockMuxMiddlewareInvocation(next http.HandlerFunc, requestID string, logger *logrus.Logger, requestPath string) *test.Hook {
	if requestPath == "" {
		requestPath = defaultRequestPath
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
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

	handler := mux.RequestGorillaMuxMiddlewareLogger(glogrus.GetLogger(logrus.NewEntry(logger)), []string{"/-/"})
	// invoke the handler
	server := handler(next)
	// Create a response writer
	writer := httptest.NewRecorder()
	// Serve HTTP server
	server.ServeHTTP(writer, req)
	return hook
}
