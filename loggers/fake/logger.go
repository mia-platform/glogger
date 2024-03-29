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
	"sync"

	"github.com/mia-platform/glogger/v4/loggers/core"
)

type Record struct {
	Fields  map[string]any
	Message string
	Level   string
	Context context.Context
}

type Entry struct {
	Logger
	fields         map[string]any
	records        []Record
	originalLogger *Logger
	ctx            context.Context
}

type Logger struct {
	mu     sync.RWMutex
	Fields map[string]any
	entry  *Entry
}

func (l *Logger) setRecord(level, msg string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	l.entry.records = append(l.entry.records, Record{
		Fields:  l.Fields,
		Message: msg,
		Level:   level,
		Context: l.entry.ctx,
	})

	if originalLogger := l.entry.originalLogger; originalLogger != nil {
		originalLogger.mu.RLock()
		defer originalLogger.mu.RUnlock()

		originalLogger.entry.records = append(originalLogger.entry.records, Record{
			Fields:  l.Fields,
			Message: msg,
			Level:   level,
			Context: l.entry.ctx,
		})
	}
}

func (l *Logger) Info(msg string) {
	l.setRecord("info", msg)
}

func (l *Logger) Trace(msg string) {
	l.setRecord("trace", msg)
}

func (l *Logger) WithFields(fields map[string]any) core.Logger[*Entry] {
	l.mu.RLock()
	defer l.mu.RUnlock()

	clonedFields := map[string]any{}
	for k, v := range l.Fields {
		clonedFields[k] = v
	}

	originalLogger := l
	if l.entry.originalLogger != nil {
		originalLogger = l.entry.originalLogger
	}

	logger := &Logger{
		Fields: clonedFields,
		entry: &Entry{
			Logger: Logger{
				Fields: clonedFields,
				entry:  l.entry,
			},
			originalLogger: originalLogger,
			records:        l.entry.records,
			ctx:            l.entry.ctx,
			fields:         clonedFields,
		},
	}
	for k, v := range fields {
		logger.entry.fields[k] = v
	}
	return logger
}

func (l *Logger) WithContext(ctx context.Context) core.Logger[*Entry] {
	l.mu.RLock()
	defer l.mu.RUnlock()

	clonedFields := map[string]any{}
	for k, v := range l.Fields {
		clonedFields[k] = v
	}

	originalLogger := l
	if l.entry.originalLogger != nil {
		originalLogger = l.entry.originalLogger
	}

	logger := &Logger{
		Fields: l.Fields,
		entry: &Entry{
			Logger: Logger{
				Fields: clonedFields,
				entry:  l.entry,
			},
			originalLogger: originalLogger,
			records:        l.entry.records,
			ctx:            ctx,
			fields:         l.Fields,
		},
	}
	return logger
}

func (e *Entry) AllRecords() []Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.records
}

func (l *Logger) OriginalLogger() *Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.entry
}

func GetLogger() core.Logger[*Entry] {
	return &Logger{
		Fields: map[string]any{},
		entry: &Entry{
			records: []Record{},
		},
	}
}
