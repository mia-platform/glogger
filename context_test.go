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

	"github.com/mia-platform/glogger/v3/loggers/core"
	"github.com/mia-platform/glogger/v3/loggers/fake"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLoggerContext(t *testing.T) {
	t.Run("create and retrieve context logger correctly", func(t *testing.T) {
		ctx := context.Background()
		loggerToSave := fake.GetLogger().WithFields(map[string]any{
			"test": "one",
		})

		ctx = WithLogger(ctx, loggerToSave)

		logger := GetOrDie[core.Logger[*fake.Entry]](ctx)
		require.Equal(t, logger.GetOriginalLogger().Fields["test"], "one")
	})

	t.Run("error if logger is not of the correct type", func(t *testing.T) {
		ctx := context.Background()
		notALogger := "something"
		ctx = context.WithValue(ctx, loggerKey{}, notALogger)

		_, err := Get[*logrus.Entry](ctx)
		require.EqualError(t, err, "logger type is not correct")
	})

	t.Run("error if logger is not in context", func(t *testing.T) {
		ctx := context.Background()

		_, err := Get[*logrus.Entry](ctx)
		require.EqualError(t, err, "logger not found in context")
	})

	t.Run("GetOrDie panic if not logger found", func(t *testing.T) {
		ctx := context.Background()
		require.PanicsWithError(t, "logger not found in context", func() { GetOrDie[*logrus.Entry](ctx) })
	})
}
