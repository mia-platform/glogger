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

package logrus

import (
	"context"

	"github.com/mia-platform/glogger/v3"
	"github.com/sirupsen/logrus"
)

var defaultLogger *logrus.Entry = logrus.NewEntry(logrus.StandardLogger())

type Logger struct {
	logger *logrus.Entry
}

func (l Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l Logger) Trace(msg string) {
	l.logger.Trace(msg)
}

func (l *Logger) WithFields(fields map[string]any) glogger.Logger[*logrus.Entry] {
	entry := l.logger.WithFields(logrus.Fields(fields))
	l.logger = entry
	return l
}

func (l Logger) GetOriginalLogger() *logrus.Entry {
	return l.logger
}

func GetLogger(logrus *logrus.Entry) glogger.Logger[*logrus.Entry] {
	return &Logger{
		logger: logrus,
	}
}

func GetFromContext(ctx context.Context) *logrus.Entry {
	entry, err := glogger.Get[*logrus.Entry](ctx)
	if err != nil {
		return defaultLogger
	}
	return entry
}
