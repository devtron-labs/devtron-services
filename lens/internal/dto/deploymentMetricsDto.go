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

package dto

import (
	"time"
)

type Metrics struct {
	Series                 []*Metric `json:"series"`
	AverageCycleTime       float64   `json:"average_cycle_time"`
	AverageLeadTime        float64   `json:"average_lead_time"`
	ChangeFailureRate      float64   `json:"change_failure_rate"`
	AverageRecoveryTime    float64   `json:"average_recovery_time"`
	AverageDeploymentSize  float32   `json:"average_deployment_size"`
	AverageLineAdded       float32   `json:"average_line_added"`
	AverageLineDeleted     float32   `json:"average_line_deleted"`
	LastFailedTime         string    `json:"last_failed_time"`
	RecoveryTimeLastFailed float64   `json:"recovery_time_last_failed"`
}

type Metric struct {
	ReleaseType           ReleaseType   `json:"release_type"`
	ReleaseStatus         ReleaseStatus `json:"release_status"`
	ReleaseTime           time.Time     `json:"release_time"`
	ChangeSizeLineAdded   int           `json:"change_size_line_added"`
	ChangeSizeLineDeleted int           `json:"change_size_line_deleted"`
	DeploymentSize        int           `json:"deployment_size"`
	CommitHash            string        `json:"commit_hash"`
	CommitTime            time.Time     `json:"commit_time"`
	LeadTime              float64       `json:"lead_time"`
	CycleTime             float64       `json:"cycle_time"`
	RecoveryTime          float64       `json:"recovery_time"`
}

type MetricRequest struct {
	AppId int    `json:"app_id"`
	EnvId int    `json:"env_id"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type AppEnvPair struct {
	AppId int `json:"appId"`
	EnvId int `json:"envId"`
}

type BulkMetricRequest struct {
	AppEnvPairs []AppEnvPair `json:"appEnvPairs"`
	From        *time.Time   `json:"from"`
	To          *time.Time   `json:"to"`
}

type AppEnvMetrics struct {
	AppId   int      `json:"appId"`
	EnvId   int      `json:"envId"`
	Metrics *Metrics `json:"metrics"`
	Error   string   `json:"error,omitempty"`
}

type BulkMetricsResponse struct {
	Results []AppEnvMetrics `json:"results"`
}

// ----------------
type ReleaseType int

const (
	Unknown ReleaseType = iota
	RollForward
	RollBack
	Patch
)

func (releaseType ReleaseType) String() string {
	return [...]string{"Unknown", "RollForward", "RollBack", "Patch"}[releaseType]
}

// --------------
type ReleaseStatus int

const (
	Success ReleaseStatus = iota
	Failure
)

func (releaseStatus ReleaseStatus) String() string {
	return [...]string{"Success", "Failure"}[releaseStatus]
}

// ------
type ProcessStage int

const (
	Init ProcessStage = iota
	ReleaseTypeDetermined
	LeadTimeFetch
)

func (ProcessStage ProcessStage) String() string {
	return [...]string{"Init", "ReleaseTypeDetermined", "LeadTimeFetch"}[ProcessStage]
}
