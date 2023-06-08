/*
 * Copyright 2023 Mia srl
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

package fake

import (
	"context"
	"fmt"

	"github.com/mia-platform/glogger/v3"
)

type Entry struct {
	Fields  map[string]any
	Message string
	Level   string
}

type Logger struct {
	fields  map[string]any
	entries []Entry
}

func (l *Logger) setEntry(level, msg string) {
	l.entries = append(l.entries, Entry{
		Fields:  l.fields,
		Message: msg,
		Level:   level,
	})
	l.fields = map[string]any{}
}

func (l *Logger) Info(msg string) {
	l.setEntry("info", msg)
}

func (l *Logger) Trace(msg string) {
	l.setEntry("trace", msg)

}

func (l *Logger) WithFields(fields map[string]any) glogger.Logger[[]Entry] {
	for k, v := range fields {
		l.fields[k] = v
	}

	return l
}

func (l Logger) GetOriginalLogger() []Entry {
	return l.entries
}

func GetLogger() glogger.Logger[[]Entry] {
	return &Logger{
		fields:  map[string]any{},
		entries: []Entry{},
	}
}

func GetFromContext(ctx context.Context) glogger.Logger[[]Entry] {
	entry, err := glogger.Get[glogger.Logger[[]Entry]](ctx)
	if err != nil {
		panic(fmt.Errorf("logger not in context"))
	}
	return entry
}
