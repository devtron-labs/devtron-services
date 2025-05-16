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

package recovery

import (
	"testing"
	"time"

	"github.com/devtron-labs/image-scanner/pkg/security"
	"go.uber.org/zap"
)

func TestRecoveryManager_StartStop(t *testing.T) {
	// Create a logger
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	// Create a config
	config := &security.ImageScanConfig{
		EnableProgressingScanCheck: true,
		RecoveryBatchSize:          10,
		RecoveryBatchDelaySeconds:  1,
		RecoveryMaxWorkers:         3,
		RecoveryStartDelaySeconds:  1,
	}

	// Create a recovery manager with nil repositories (we're just testing start/stop)
	rm := NewRecoveryManager(
		sugaredLogger,
		config,
		nil,
		nil,
		nil,
		nil,
	)

	// Start the recovery manager
	rm.Start()

	// Check that it's running
	metrics := rm.GetMetrics()
	if !metrics.IsRunning {
		t.Error("Recovery manager should be running after Start()")
	}

	// Stop the recovery manager
	rm.Stop()

	// Wait a bit for the goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Check that it's stopped
	metrics = rm.GetMetrics()
	if metrics.IsRunning {
		t.Error("Recovery manager should not be running after Stop()")
	}
}
