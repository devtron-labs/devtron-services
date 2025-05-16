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

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devtron-labs/image-scanner/pkg/recovery"
	"github.com/devtron-labs/image-scanner/pkg/security"
	"go.uber.org/zap"
)

func TestRecoveryHandler(t *testing.T) {
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

	// Create a recovery manager
	rm := recovery.NewRecoveryManager(
		sugaredLogger,
		config,
		nil,
		nil,
		nil,
		nil,
	)

	// Create a recovery handler
	handler := NewRecoveryHandlerImpl(sugaredLogger, rm)

	// Test GetRecoveryStatus
	t.Run("GetRecoveryStatus", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/recovery/status", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlerFunc := http.HandlerFunc(handler.GetRecoveryStatus)
		handlerFunc.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var response struct {
			Code   int                      `json:"code"`
			Status string                   `json:"status"`
			Result recovery.RecoveryMetrics `json:"result"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if response.Code != 200 {
			t.Errorf("expected response code 200, got %d", response.Code)
		}
	})

	// Test StartRecovery
	t.Run("StartRecovery", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/recovery/start", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlerFunc := http.HandlerFunc(handler.StartRecovery)
		handlerFunc.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check that recovery is running
		metrics := rm.GetMetrics()
		if !metrics.IsRunning {
			t.Error("Recovery should be running after StartRecovery")
		}
	})

	// Test StopRecovery
	t.Run("StopRecovery", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/recovery/stop", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlerFunc := http.HandlerFunc(handler.StopRecovery)
		handlerFunc.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Wait a bit for the goroutine to stop
		time.Sleep(100 * time.Millisecond)

		// Check that recovery is stopped
		metrics := rm.GetMetrics()
		if metrics.IsRunning {
			t.Error("Recovery should be stopped after StopRecovery")
		}
	})
}
