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
	"net/http"

	"github.com/devtron-labs/image-scanner/pkg/recovery"
	"go.uber.org/zap"
)

type RecoveryHandler interface {
	GetRecoveryStatus(w http.ResponseWriter, r *http.Request)
	StartRecovery(w http.ResponseWriter, r *http.Request)
	StopRecovery(w http.ResponseWriter, r *http.Request)
}

type RecoveryHandlerImpl struct {
	Logger          *zap.SugaredLogger
	RecoveryManager *recovery.RecoveryManager
}

func NewRecoveryHandlerImpl(logger *zap.SugaredLogger, recoveryManager *recovery.RecoveryManager) *RecoveryHandlerImpl {
	return &RecoveryHandlerImpl{
		Logger:          logger,
		RecoveryManager: recoveryManager,
	}
}

func (impl *RecoveryHandlerImpl) GetRecoveryStatus(w http.ResponseWriter, r *http.Request) {
	metrics := impl.RecoveryManager.GetMetrics()
	WriteJsonResp(w, nil, metrics, http.StatusOK)
}

func (impl *RecoveryHandlerImpl) StartRecovery(w http.ResponseWriter, r *http.Request) {
	impl.RecoveryManager.Start()
	WriteJsonResp(w, nil, map[string]string{"status": "recovery process started"}, http.StatusOK)
}

func (impl *RecoveryHandlerImpl) StopRecovery(w http.ResponseWriter, r *http.Request) {
	impl.RecoveryManager.Stop()
	WriteJsonResp(w, nil, map[string]string{"status": "recovery process stopped"}, http.StatusOK)
}
