/*
 * Copyright (c) 2024. Devtron Inc.
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

package utils

import (
	"go.uber.org/zap"
	"time"
)

func LogExecutionTime(logger *zap.SugaredLogger, startTime time.Time, message string, keysAndValues ...interface{}) {
	if len(keysAndValues) != 0 {
		logger.Debugw(message, "time", time.Since(startTime), keysAndValues)
	} else {
		logger.Debugw(message, "time", time.Since(startTime))
	}
}
