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

package glogger

import (
	"net/http"
	"testing"
)

type ResponseWriterMock struct {
	writer            http.ResponseWriter
	headerCalled      bool
	writeHeaderCalled bool
	writeCalled       bool
}

func (r *ResponseWriterMock) Header() http.Header {
	r.headerCalled = true
	return http.Header{}
}

func (r *ResponseWriterMock) WriteHeader(status int) {
	r.writeHeaderCalled = true
}

func (r *ResponseWriterMock) Write(b []byte) (int, error) {
	r.writeCalled = true
	return 1, nil
}

func TestReadableResponseWriter(t *testing.T) {
	mock := ResponseWriterMock{}
	myw := ReadableResponseWriter{writer: &mock}

	myw.WriteHeader(200)
	if !mock.writeHeaderCalled {
		t.Errorf("mock write header not called")
	}

	body := []byte("ciao")
	_, err := myw.Write(body)
	if err != nil {
		t.Errorf("Error write body %v", err)
	}
	if !mock.writeCalled {
		t.Errorf("mock write not called")
	}

	myw.Header()
	if !mock.headerCalled {
		t.Errorf("mock header not called")
	}
}
