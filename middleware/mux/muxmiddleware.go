package mux

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mia-platform/glogger/v3"
	"github.com/mia-platform/glogger/v3/loggers/core"
	"github.com/mia-platform/glogger/v3/middleware/utils"
)

type muxLoggingContext struct {
	req *http.Request
	res *readableResponseWriter
}

func (mlc *muxLoggingContext) Request() glogger.RequestLoggingContext {
	return &muxRequestLoggingContext{mlc.req}
}

func (mlc *muxLoggingContext) Response() glogger.ResponseLoggingContext {
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
	return mrlc.req.Method
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
	return mrlc.res.StatusCode
}

// RequestMiddlewareLogger is a gorilla/mux middleware to log all requests
// It logs the incoming request and when request is completed, adding latency of the request
func RequestMiddlewareLogger[Logger any](logger core.Logger[Logger], excludedPrefix []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			myw := readableResponseWriter{Writer: w, StatusCode: http.StatusOK}
			muxLoggingContext := &muxLoggingContext{
				req: r,
				res: &myw,
			}

			requestID := utils.GetReqID(muxLoggingContext)
			loggerWithReqId := logger.WithFields(map[string]any{
				"reqId": requestID,
			})
			ctx := glogger.WithLogger(r.Context(), loggerWithReqId.OriginalLogger())

			// Skip logging for excluded routes
			for _, prefix := range excludedPrefix {
				if strings.HasPrefix(r.URL.RequestURI(), prefix) {
					next.ServeHTTP(&myw, r.WithContext(ctx))
					return
				}
			}

			utils.LogIncomingRequest[Logger](muxLoggingContext, loggerWithReqId)
			next.ServeHTTP(&myw, r.WithContext(ctx))
			utils.LogRequestCompleted[Logger](muxLoggingContext, loggerWithReqId, start)
		})
	}
}
