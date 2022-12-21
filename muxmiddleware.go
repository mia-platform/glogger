package glogger

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type muxLoggingContext struct {
	req *http.Request
	res *readableResponseWriter
	ctx context.Context
}

func (mlc *muxLoggingContext) Context() context.Context {
	return mlc.ctx
}

func (mlc *muxLoggingContext) Request() RequestLoggingContext {
	return &muxRequestLoggingContext{mlc.req}
}

func (mlc *muxLoggingContext) Response() ResponseLoggingContext {
	return &muxResponseLoggingContext{mlc.res}
}

type muxRequestLoggingContext struct {
	req *http.Request
}

func (mrlc *muxRequestLoggingContext) GetHeader(key string) string {
	return mrlc.req.Header.Get(key)
}

func (mrlc *muxRequestLoggingContext) URI() string {
	return mrlc.req.URL.RequestURI()
}

func (mrlc *muxRequestLoggingContext) Host() string {
	return mrlc.req.Host
}

func (mrlc *muxRequestLoggingContext) Method() string {
	return mrlc.req.Host
}

type muxResponseLoggingContext struct {
	res *readableResponseWriter
}

func (mrlc *muxResponseLoggingContext) BodySize() int {
	if content := mrlc.res.Header().Get("Content-Length"); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return mrlc.res.Length()
}

func (mrlc *muxResponseLoggingContext) StatusCode() int {
	return mrlc.res.statusCode
}

// RequestMuxMiddlewareLogger is a gorilla/mux middleware to log all requests with logrus
// It logs the incoming request and when request is completed, adding latency of the request
func RequestMuxMiddlewareLogger(logger *logrus.Logger, excludedPrefix []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := getReqID(logger, r.Header.Get)
			ctx := WithLogger(r.Context(), logrus.NewEntry(logger).WithFields(logrus.Fields{
				"reqId": requestID,
			}))
			myw := readableResponseWriter{writer: w, statusCode: http.StatusOK}

			// Skip logging for excluded routes
			for _, prefix := range excludedPrefix {
				if strings.HasPrefix(r.URL.RequestURI(), prefix) {
					next.ServeHTTP(&myw, r.WithContext(ctx))
					return
				}
			}

			muxLoggingContext := &muxLoggingContext{
				ctx: ctx,
				req: r,
				res: &myw,
			}

			logBeforeHandler(muxLoggingContext)
			next.ServeHTTP(&myw, r.WithContext(ctx))
			logAfterHandler(muxLoggingContext, start)
		})
	}
}
