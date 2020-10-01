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
	"github.com/sirupsen/logrus"
)

// InitOptions is the struct of options to configure the logger
type InitOptions struct {
	Level string
}

// InitHelper is a function to init json logger
func InitHelper(options InitOptions) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(&CustomWriter{})
	if options.Level == "" {
		return logger, nil
	}
	level, err := logrus.ParseLevel(options.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)
	return logger, nil
}
