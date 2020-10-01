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
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestCustomWriter(t *testing.T) {
	t.Run("_", func(t *testing.T) {

		now := time.Now()
		logEntryMap := map[string]interface{}{
			"level": "info",
			"time":  now.Format(time.RFC3339),
		}

		logEntry, err := json.Marshal(logEntryMap)
		assert.Assert(t, err == nil, "failed log entry map serialization: %s", err)

		c := CustomWriter{}
		n, err := c.Write(logEntry)
		assert.Assert(t, err == nil, "failed custom writer writing: %s", err)
		assert.Assert(t, n == 30, "invalid length returned, found: %d", n)
	})
}
