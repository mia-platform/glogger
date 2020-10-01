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
	"fmt"
	"os"
	"time"
)

type CustomWriter struct {
}

type LogEntry map[string]interface{}

func (c *CustomWriter) Write(msg []byte) (int, error) {
	// fmt.Printf("CUSTOM WRITE>>>>>>>>>>> %s\n", string(msg))

	// var re = regexp.MustCompile(`"level":"info"`)
	// s := re.ReplaceAllString(string(msg), `"level":20`)
	// fmt.Printf("REPLACED msg >>> %s\n", s)

	logEntry := make(LogEntry)

	if err := json.Unmarshal(msg, &logEntry); err != nil {
		return 0, err
	}

	switch logEntry["level"] {
	case "info":
		logEntry["level"] = 20
	}

	entryTime, ok := logEntry["time"].(string)

	if !ok {
		return 0, fmt.Errorf("log entry time is not a string")
	}

	logTime, err := time.Parse(time.RFC3339, entryTime)
	if err != nil {
		return 0, err
	}

	logEntry["time"] = logTime.Unix()
	s, err := json.Marshal(logEntry)
	if err != nil {
		return 0, err
	}

	return os.Stdout.Write([]byte(s))
}
