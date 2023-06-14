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

package logrus

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestInitHelper(t *testing.T) {
	t.Run("if LOG_LEVEL not defined, return logger with info value", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{})

		require.True(t, err == nil, "Error setting default logger value")
		require.Equal(t, logrus.InfoLevel, logger.GetLevel(), "Default level value")
	})

	t.Run("level correctly set from env variable", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{Level: "warn"})

		require.True(t, err == nil, "Error setting default logger value")
		require.Equal(t, logrus.WarnLevel, logger.GetLevel(), "Set log level value from env variable")
	})

	t.Run("set an invalid level from env variable return error", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{Level: "not a real level"})

		require.True(t, logger == nil, "Logger is nil.")
		require.True(t, err != nil, "An error is expected. Found nil instead.")
	})

	t.Run("custom JSONFormatter integration", func(t *testing.T) {
		now := time.Now()
		var buffer bytes.Buffer

		logger, _ := InitHelper(InitOptions{})
		logger.Out = &buffer
		logger.WithTime(now)
		logger.WithField("foo", "bar").Info("hello")

		type log struct {
			Level   int    `json:"level"`
			Message string `json:"msg"`
			Foo     string `json:"foo"`
			Time    int64  `json:"time"`
		}

		var result log
		err := json.Unmarshal(buffer.Bytes(), &result)

		require.True(t, err == nil, "Unexpected Error %s", err)
		require.Equal(t, log{
			Level:   30,
			Message: "hello",
			Foo:     "bar",
			Time:    now.UnixNano() / int64(1e6),
		}, result)
	})
}
