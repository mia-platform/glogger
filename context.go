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
	"context"
	"fmt"
)

type loggerKey struct{}

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger[Logger any](ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Get retrieves the current logger from the context.
func Get[Logger any](ctx context.Context) (Logger, error) {
	loggerFromCtx := ctx.Value(loggerKey{})

	if loggerFromCtx == nil {
		return *new(Logger), fmt.Errorf("logger not found in context")
	}

	logger, ok := loggerFromCtx.(Logger)
	if !ok {
		return *new(Logger), fmt.Errorf("logger type is not correct")
	}
	return logger, nil
}

// Get retrieves the current logger from the context.
func GetOrDie[Logger any](ctx context.Context) Logger {
	logger, err := Get[Logger](ctx)
	if err != nil {
		panic(err)
	}
	return logger
}
