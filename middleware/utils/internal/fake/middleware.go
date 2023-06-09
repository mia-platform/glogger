package fake

import (
	"context"

	"github.com/mia-platform/glogger/v3"
)

type fakeLoggingContext struct {
	req Request
	res Response
	ctx context.Context
}

type Request struct {
	Headers map[string]string
	URI     string
}

type Response struct {
	StatusCode int
	BodySize   int
}

func NewContext(ctx context.Context, req Request, res Response) glogger.LoggingContext {
	return &fakeLoggingContext{
		req: req,
		res: res,
		ctx: ctx,
	}
}

func (flc *fakeLoggingContext) Context() context.Context {
	return flc.ctx
}

func (flc *fakeLoggingContext) Request() glogger.RequestLoggingContext {
	return flc
}

func (flc *fakeLoggingContext) Response() glogger.ResponseLoggingContext {
	return flc
}

func (flc *fakeLoggingContext) GetHeader(key string) string {
	return flc.req.Headers[key]
}

func (flc *fakeLoggingContext) URI() string {
	return "/custom-uri"
}

func (flc *fakeLoggingContext) Host() string {
	return "echo-service:3456"
}

func (flc *fakeLoggingContext) Method() string {
	return "GET"
}

func (flc *fakeLoggingContext) BodySize() int {
	return flc.res.BodySize
}

func (flc *fakeLoggingContext) StatusCode() int {
	return flc.res.StatusCode
}
