/*
 * Copyright 2020 Mia srl
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
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestCustomWriter(t *testing.T) {
	t.Run("level transformer", func(t *testing.T) {
		now := time.Now()
		testCases := []struct {
			inputLevel    logrus.Level
			expectedLevel int
			inputTime     time.Time
			expectedTime  int64
		}{
			{
				inputLevel:    logrus.TraceLevel,
				expectedLevel: 10,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.DebugLevel,
				expectedLevel: 20,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.InfoLevel,
				expectedLevel: 30,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.WarnLevel,
				expectedLevel: 40,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.ErrorLevel,
				expectedLevel: 50,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.FatalLevel,
				expectedLevel: 60,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
			{
				inputLevel:    logrus.PanicLevel,
				expectedLevel: 70,
				inputTime:     now,
				expectedTime:  now.Unix(),
			},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("test case for level %s", testCase.inputLevel), func(t *testing.T) {
				logEntry := logrus.Entry{
					Level:   testCase.inputLevel,
					Time:    testCase.inputTime,
					Message: "test",
					Data:    logrus.Fields{},
				}
				c := JSONFormatter{}

				result, err := c.Format(&logEntry)

				assert.Assert(t, err == nil, "failed custom writer writing: %s", err)
				assert.Equal(t, string(result), fmt.Sprintf("{\"level\":%d,\"msg\":\"test\",\"time\":%d}\n", testCase.expectedLevel, testCase.expectedTime))

			})
		}
	})
}
