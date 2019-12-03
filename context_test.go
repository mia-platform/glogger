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
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gotest.tools/assert"
)

func TestLoggerContext(t *testing.T) {
	t.Run("create and retrieve context logger correctly", func(t *testing.T) {
		ctx := context.Background()

		ctx = WithLogger(ctx, Get(ctx).WithField("test", "one"))
		assert.Equal(t, Get(ctx).Data["test"], "one")
	})

	t.Run("create and retrieve context logger correctly with correct time", func(t *testing.T) {
		ctx := context.Background()

		ctx = WithLogger(ctx, Get(ctx).WithField("test", "one"))
		assert.Equal(t, Get(ctx).Data["test"], "one")
	})

	t.Run("if logger is not of the correct type, return new logger", func(t *testing.T) {
		ctx := context.Background()
		notALogger := "something"
		ctx = context.WithValue(ctx, loggerKey{}, notALogger)

		logFromContext := Get(ctx)

		assert.Assert(t, logFromContext != nil, "log from context does not panic")
	})

	t.Run("if logger is not of the correct type, return new logger with correct time", func(t *testing.T) {
		ctx := context.Background()
		notALogger := "something"
		ctx = context.WithValue(ctx, loggerKey{}, notALogger)

		logFromContext := Get(ctx)

		assert.Assert(t, logFromContext != nil, "log from context does not panic")
	})

	t.Run("if logger is overwrite in context, type conversion does not panic", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := test.NewNullLogger()
		ctx = WithLogger(ctx, logrus.NewEntry(logger))
		notALogger := "something"
		ctx = context.WithValue(ctx, loggerKey{}, notALogger)

		logFromContext := Get(ctx)
		var hasPanic = false
		if r := recover(); r != nil {
			hasPanic = true
		}
		assert.Assert(t, logFromContext != nil, "log from context does not panic")
		assert.Assert(t, !hasPanic, "Get log from context panic")
	})
}
