/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package logger

import (
	"github.com/caarlos0/env"
	internalotel "github.com/devtron-labs/git-sensor/internals/otel"
	otelzap "go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level int `env:"LOG_LEVEL" envDefault:"0"` // default info
}

func NewSugaredLogger() *zap.SugaredLogger {
	logConfig := &LogConfig{}
	err := env.Parse(logConfig)
	if err != nil {
		panic("failed to parse env config for logger: " + err.Error())
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(logConfig.Level))

	log, err := config.Build()
	if err != nil {
		panic("failed to create the logger: " + err.Error())
	}

	// Bridge zap logs to the OTel log provider when OTel is enabled.
	if internalotel.Enabled && internalotel.LP != nil {
		otelCore := otelzap.NewCore("git-sensor",
			otelzap.WithLoggerProvider(internalotel.LP),
		)
		log = log.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, otelCore)
		}))
	}

	return log.Sugar()
}
