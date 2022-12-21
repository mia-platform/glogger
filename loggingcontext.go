package glogger

import "context"

type LoggingContext interface {
	Context() context.Context
	Request() RequestLoggingContext
	Response() ResponseLoggingContext
}

type RequestLoggingContext interface {
	GetHeader(string) string
	URI() string
	Host() string
	Method() string
}

type ResponseLoggingContext interface {
	BodySize() int
	StatusCode() int
}
