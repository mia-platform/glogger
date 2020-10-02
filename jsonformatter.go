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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type JSONFormatter struct {
	// DisableHTMLEscape allows disabling html escaping in output
	DisableHTMLEscape bool

	// PrettyPrint will indent all json logs
	PrettyPrint bool
}

func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	data["time"] = entry.Time.Unix()
	data["msg"] = entry.Message
	data["level"] = getLevelFromString(entry.Level)

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!f.DisableHTMLEscape)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return b.Bytes(), nil
}

func getLevelFromString(logLevel logrus.Level) int {
	switch logLevel {
	case logrus.TraceLevel:
		return 10
	case logrus.DebugLevel:
		return 20
	case logrus.InfoLevel:
		return 30
	case logrus.WarnLevel:
		return 40
	case logrus.ErrorLevel:
		return 50
	case logrus.FatalLevel:
		return 60
	case logrus.PanicLevel:
		return 70
	default:
		return 30
	}
}
