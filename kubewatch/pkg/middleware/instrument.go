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

package middleware

import (
	"github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

// metrics name constants
const (
	KUBEWATCH_UNREACHABLE_CLIENT_COUNT        = "Kubewatch_unreachable_client_count"
	KUBEWATCH_UNREGISTERED_INFORMER_COUNT     = "Kubewatch_unregistered_informer_count"
	KUBEWATCH_NON_ADMINISTRATIVE_EVENTS_COUNT = "Kubewatch_non_administrative_events_count"
)

// metrics labels constants
const (
	CLUSTER_NAME  = "clusterName"
	CLUSTER_ID    = "clusterId"
	INFORMER_NAME = "informerName"
	RESOURCE_TYPE = "resourceType"
)

// ResourceMetrics type
type ResourceMetrics string

func (m ResourceMetrics) String() string {
	return string(m)
}

// resource labels values constants
const (
	RESOURCE_ARGO_WORKFLOW ResourceMetrics = "ArgoWorkflow"
	RESOURCE_K8S_JOB       ResourceMetrics = "K8sJob"
)

// InformerMetrics type
type InformerMetrics string

func (m InformerMetrics) String() string {
	return string(m)
}

// informer labels values constants
const (
	CI_STAGE_ARGO_WORKFLOW_INFORMER InformerMetrics = "CIStageArgoWorkflow"
	CD_STAGE_ARGO_WORLFLOW_INFORMER InformerMetrics = "CDStageArgoWorkflow"
	ARGO_CD_INFORMER                InformerMetrics = "ArgoCD"
	DEFAULT_CLUSTER_SECRET_INFORMER InformerMetrics = "DefaultClusterSecret"
	SYSTEM_EXECUTOR_INFORMER        InformerMetrics = "SystemExecutor"
)

var UnreachableCluster = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: KUBEWATCH_UNREACHABLE_CLIENT_COUNT,
		Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
	},
	[]string{CLUSTER_NAME, CLUSTER_ID},
)

var UnregisteredInformers = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: KUBEWATCH_UNREGISTERED_INFORMER_COUNT,
		Help: "How many informers are unregistered, with informer name, cluster name and cluster id.",
	},
	[]string{CLUSTER_NAME, CLUSTER_ID, INFORMER_NAME},
)

var NonAdministrativeEvents = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: KUBEWATCH_NON_ADMINISTRATIVE_EVENTS_COUNT,
		Help: "How many events are non-administrative (not devtron managed), with resource type, cluster name and cluster id.",
	},
	[]string{CLUSTER_NAME, CLUSTER_ID, RESOURCE_TYPE},
)

func IncUnreachableCluster(clusterLabels *bean.ClusterLabels) {
	UnreachableCluster.WithLabelValues(clusterLabels.ClusterName, strconv.Itoa(clusterLabels.ClusterId)).Inc()
}

func IncUnregisteredInformers(clusterLabels *bean.ClusterLabels, informerName InformerMetrics) {
	UnregisteredInformers.WithLabelValues(clusterLabels.ClusterName, strconv.Itoa(clusterLabels.ClusterId), informerName.String()).Inc()
}

func IncNonAdministrativeEvents(clusterLabels *bean.ClusterLabels, resourceType ResourceMetrics) {
	NonAdministrativeEvents.WithLabelValues(clusterLabels.ClusterName, strconv.Itoa(clusterLabels.ClusterId), resourceType.String()).Inc()
}
