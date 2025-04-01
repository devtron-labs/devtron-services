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

package bean

import "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

type ClusterInfo struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	BearerToken string `json:"bearerToken"`
	ServerUrl   string `json:"serverUrl"`
}

const (
	ClusterModifyEventFieldSelector = "type==cluster.request/modify"
	InformerAlreadyExistMessage     = "INFORMER_ALREADY_EXIST"

	ExitCode143Error   = "Error (exit code 143)"
	NodeNoLongerExists = "PodGC: node no longer exists"
	UpdateEvent        = "update_event"
	DeleteEvent        = "delete_event"

	WorkflowLabelSelector = "devtron.ai/purpose==workflow"
	JobKind               = "Job"
)

type ClusterLabels struct {
	ClusterName string
	ClusterId   int
}

func NewClusterLabels(clusterName string, clusterId int) *ClusterLabels {
	return &ClusterLabels{
		ClusterName: clusterName,
		ClusterId:   clusterId,
	}
}

const (
	DevtronAdministratorInstance = "devtronAdministratorInstance"
)

type CiCdStatus struct {
	DevtronAdministratorInstance string `json:"devtronAdministratorInstance"`
	*v1alpha1.WorkflowStatus
}
