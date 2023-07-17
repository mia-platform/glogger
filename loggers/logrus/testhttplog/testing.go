package testhttplog

import (
	"testing"

	"github.com/mia-platform/glogger/v4/middleware/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type ExpectedLogFields struct {
	Level     logrus.Level
	RequestID string
	Message   string
}

type ExpectedIncomingLogFields struct {
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
}

type ExpectedOutcomingLogFields struct {
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
	Bytes         int
	StatusCode    int
}

func LogAssertions(t *testing.T, logEntry *logrus.Entry, reqIDKey string, expected ExpectedLogFields) string {
	t.Helper()
	require.Equal(t, expected.Level, logEntry.Level, "Unexpected level of log for log in incoming request")
	require.Equal(t, expected.Message, logEntry.Message, "Unexpected message of log for log in incoming request")
	requestID := logEntry.Data[reqIDKey]
	_, ok := requestID.(string)
	require.True(t, ok, "Unexpected or empty requestID for log in incoming request")
	if expected.RequestID != "" {
		require.Equal(t, expected.RequestID, requestID, "Unexpected requestID for log in incoming request")
	}
	return requestID.(string)
}

func IncomingRequestAssertions(t *testing.T, incomingRequestLogEntry *logrus.Entry, expected ExpectedIncomingLogFields) {
	http := incomingRequestLogEntry.Data["http"].(utils.HTTP)
	require.Equal(t, expected.Method, http.Request.Method, "Unexpected http method for log in incoming request")
	require.Equal(t, expected.Original, http.Request.UserAgent.Original, "Unexpected original userAgent for log of request completed")

	url := incomingRequestLogEntry.Data["url"].(utils.URL)
	require.Equal(t, expected.Path, url.Path, "Unexpected http uri path for log in incoming request")

	host := incomingRequestLogEntry.Data["host"].(utils.Host)
	require.Equal(t, expected.Hostname, host.Hostname, "Unexpected hostname for log of request completed")
	require.Equal(t, expected.ForwardedHost, host.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	require.Equal(t, expected.IP, host.IP, "Unexpected ip for log of request completed")
}

func OutgoingRequestAssertions(t *testing.T, outcomingRequestLogEntry *logrus.Entry, expected ExpectedOutcomingLogFields) {
	http := outcomingRequestLogEntry.Data["http"].(utils.HTTP)
	require.Equal(t, expected.Method, http.Request.Method, "Unexpected http method for log in incoming request")
	require.Equal(t, expected.Original, http.Request.UserAgent.Original, "Unexpected original userAgent for log of request completed")
	require.Equal(t, expected.StatusCode, http.Response.StatusCode, "Unexpected status code for log of request completed")
	require.Equal(t, expected.Bytes, http.Response.Body.Bytes, "Unexpected status code for log of request completed")

	url := outcomingRequestLogEntry.Data["url"].(utils.URL)
	require.Equal(t, expected.Path, url.Path, "Unexpected http uri path for log in incoming request")

	host := outcomingRequestLogEntry.Data["host"].(utils.Host)
	require.Equal(t, expected.Hostname, host.Hostname, "Unexpected hostname for log of request completed")
	require.Equal(t, expected.ForwardedHost, host.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	require.Equal(t, expected.IP, host.IP, "Unexpected ip for log of request completed")

	_, ok := outcomingRequestLogEntry.Data["responseTime"].(float64)
	require.True(t, ok, "Invalid took duration for log of request completed")
}
