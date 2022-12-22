package middleware

import "context"

type loggingContext interface {
	Context() context.Context
	Request() requestLoggingContext
	Response() responseLoggingContext
}

type requestLoggingContext interface {
	GetHeader(string) string
	URI() string
	Host() string
	Method() string
}

type responseLoggingContext interface {
	BodySize() int
	StatusCode() int
}
