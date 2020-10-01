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

// func TestCustomWriter(t *testing.T) {
// 	t.Run("level transformer", func(t *testing.T) {
// 		now := time.Now()
// 		testCases := []struct {
// 			inputLevel string
// 			inputTime  string
// 			expected   string
// 		}{
// 			{
// 				inputLevel: "trace",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":10,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "debug",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":20,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "info",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":30,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "warn",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":40,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "error",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":50,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "fatal",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":60,"time":%d}`, now.Unix()),
// 			},
// 			{
// 				inputLevel: "panic",
// 				inputTime:  now.Format(time.RFC3339),
// 				expected:   fmt.Sprintf(`{"level":70,"time":%d}`, now.Unix()),
// 			},
// 		}

// 		for _, testCase := range testCases {
// 			t.Run(fmt.Sprintf("test case for level %s", testCase.inputLevel), func(t *testing.T) {
// 				logEntryMap := map[string]interface{}{
// 					"level": testCase.inputLevel,
// 					"time":  now.Format(time.RFC3339),
// 				}

// 				logEntry, err := json.Marshal(logEntryMap)
// 				assert.Assert(t, err == nil, "failed log entry map serialization: %s", err)

// 				c := CustomWriter{}
// 				var n int
// 				result := captureStdout(t, func() {
// 					n, err = c.Write(logEntry)
// 				})
// 				assert.Assert(t, err == nil, "failed custom writer writing: %s", err)
// 				assert.Assert(t, n == 30, "invalid length returned, found: %d", n)
// 				assert.Equal(t, result, testCase.expected, now.Unix())
// 			})
// 		}
// 	})
// }
