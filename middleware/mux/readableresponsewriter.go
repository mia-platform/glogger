/*
 * Copyright 2019 Mia srl
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mux

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// readableResponseWriter struct, add readable statusCode to ResponseWriter
type readableResponseWriter struct {
	Writer     http.ResponseWriter
	StatusCode int
	length     int
}

// WriteHeader func, set statusCode parameter
func (r *readableResponseWriter) WriteHeader(code int) {
	r.StatusCode = code
	r.Writer.WriteHeader(code)
}

// Write func, calls ResponseWriter Write fn
func (r *readableResponseWriter) Write(b []byte) (int, error) {
	n, err := r.Writer.Write(b)

	if err != nil {
		return n, err
	}

	r.length += n
	return n, err
}

// Header func, calls ResponseWriter Header fn
func (r *readableResponseWriter) Header() http.Header {
	return r.Writer.Header()
}

func (r *readableResponseWriter) Length() int {
	return r.length
}

// Flush to implement http.Flusher interface
func (r *readableResponseWriter) Flush() {
	if flusherWriter, ok := r.Writer.(http.Flusher); ok {
		flusherWriter.Flush()
	}
}

// Hijack func, calls the underlying ResponseWriter Hijack fn
func (r *readableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.Writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return nil, nil, fmt.Errorf("the Hijacker interface is not supported")
}
