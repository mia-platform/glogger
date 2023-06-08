package testhttplog

import (
	"testing"

	"github.com/mia-platform/glogger/v3/middleware/core"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
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
	assert.Equal(t, logEntry.Level, expected.Level, "Unexpected level of log for log in incoming request")
	assert.Equal(t, logEntry.Message, expected.Message, "Unexpected message of log for log in incoming request")
	requestID := logEntry.Data[reqIDKey]
	_, ok := requestID.(string)
	assert.Assert(t, ok, "Unexpected or empty requestID for log in incoming request")
	if expected.RequestID != "" {
		assert.Equal(t, requestID, expected.RequestID, "Unexpected requestID for log in incoming request")
	}
	return requestID.(string)
}

func IncomingRequestAssertions(t *testing.T, incomingRequestLogEntry *logrus.Entry, expected ExpectedIncomingLogFields) {
	http := incomingRequestLogEntry.Data["http"].(core.HTTP)
	assert.Equal(t, http.Request.Method, expected.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original, "Unexpected original userAgent for log of request completed")

	url := incomingRequestLogEntry.Data["url"].(core.URL)
	assert.Equal(t, url.Path, expected.Path, "Unexpected http uri path for log in incoming request")

	host := incomingRequestLogEntry.Data["host"].(core.Host)
	assert.Equal(t, host.Hostname, expected.Hostname, "Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP, "Unexpected ip for log of request completed")
}

func OutgoingRequestAssertions(t *testing.T, outcomingRequestLogEntry *logrus.Entry, expected ExpectedOutcomingLogFields) {
	http := outcomingRequestLogEntry.Data["http"].(core.HTTP)
	assert.Equal(t, http.Request.Method, expected.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original, "Unexpected original userAgent for log of request completed")
	assert.Equal(t, http.Response.StatusCode, expected.StatusCode, "Unexpected status code for log of request completed")
	assert.Equal(t, http.Response.Body["bytes"], expected.Bytes, "Unexpected status code for log of request completed")

	url := outcomingRequestLogEntry.Data["url"].(core.URL)
	assert.Equal(t, url.Path, expected.Path, "Unexpected http uri path for log in incoming request")

	host := outcomingRequestLogEntry.Data["host"].(core.Host)
	assert.Equal(t, host.Hostname, expected.Hostname, "Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost, "Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP, "Unexpected ip for log of request completed")

	_, ok := outcomingRequestLogEntry.Data["responseTime"].(float64)
	assert.Assert(t, ok, "Invalid took duration for log of request completed")
}
