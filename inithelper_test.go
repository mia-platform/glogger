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
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

const logLevelEnv = "LOG_LEVEL"

func TestInitHelper(t *testing.T) {
	t.Run("if LOG_LEVEL not defined, return logger with info value", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{})

		assert.Assert(t, err == nil, "Error setting default logger value")
		assert.Equal(t, logger.GetLevel(), logrus.InfoLevel, "Default level value")
	})

	t.Run("level correctly set from env variable", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{Level: "warn"})

		assert.Assert(t, err == nil, "Error setting default logger value")
		assert.Equal(t, logger.GetLevel(), logrus.WarnLevel, "Set log level value from env variable")
	})

	t.Run("set an invalid level from env variable return error", func(t *testing.T) {
		logger, err := InitHelper(InitOptions{Level: "not a real level"})

		assert.Assert(t, logger == nil, "Logger is nil.")
		assert.Assert(t, err != nil, "An error is expected. Found nil instead.")
	})

	t.Run("customWriter integration", func(t *testing.T) {
		originalStdout := os.Stdout
		// defer func() {

		// }()

		r, w, _ := os.Pipe()

		// fmt.Printf("error %s\n", err)

		os.Stdout = w
		logger, _ := InitHelper(InitOptions{})
		now := time.Now()
		logger.WithTime(now)
		logger.Info("ciao")
		w.Close()
		bytes, _ := ioutil.ReadAll(r)
		// fmt.Printf("error2 %s\n", err2)
		r.Close()
		os.Stdout = originalStdout
		fmt.Printf("test %s\n", string(bytes))

		assert.Equal(t, string(bytes), fmt.Sprintf(`{"level":20,"msg":"ciao","time":%d}`, now.Unix()))
	})
}
