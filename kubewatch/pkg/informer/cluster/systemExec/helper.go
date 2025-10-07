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

package systemExec

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	informerBean "github.com/devtron-labs/common-lib/informer"
	"github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"golang.org/x/exp/maps"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sort"
)

func (impl *InformerImpl) getSystemWfStopper(clusterId int) (*bean.FactoryStopper, bool) {
	stopper, ok := impl.systemWfInformerStopper[clusterId]
	if ok {
		return stopper, stopper.HasInformer()
	}
	return stopper, false
}

func (impl *InformerImpl) getStoppableClusterIds() []int {
	return maps.Keys(impl.systemWfInformerStopper)
}

func (impl *InformerImpl) getStopChannel(informerFactory kubeinformers.SharedInformerFactory, clusterLabels *bean.ClusterLabels) (chan struct{}, error) {
	stopChannel := make(chan struct{})
	stopper, ok := impl.systemWfInformerStopper[clusterLabels.ClusterId]
	if ok && stopper.HasInformer() {
		impl.logger.Debug(fmt.Sprintf("system executor informer for %s already exist", clusterLabels.ClusterName))
		return stopChannel, errors.AlreadyExists
	}
	stopper = stopper.GetStopper(informerFactory, stopChannel)
	impl.systemWfInformerStopper[clusterLabels.ClusterId] = stopper
	return stopChannel, nil
}

func (impl *InformerImpl) checkIfPodDeletedAndUpdateMessage(podName, namespace string,
	nodeStatus v1alpha1.NodeStatus, config *rest.Config) (v1alpha1.NodeStatus, bool) {

	if (nodeStatus.Phase == v1alpha1.NodeFailed || nodeStatus.Phase == v1alpha1.NodeError) && (nodeStatus.Message == bean.ExitCode143Error || nodeStatus.Message == bean.NodeNoLongerExists ||
		nodeStatus.Message == bean.NodeForceDeleted) {
		clusterClient, k8sErr := impl.k8sUtil.GetK8sClientForConfig(config)
		if k8sErr != nil {
			return nodeStatus, false
		}
		pod, err := clusterClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting pod from clusterClient", "podName", podName, "namespace", namespace, "err", err)
			if isResourceNotFoundErr(err) {
				nodeStatus.Message = informerBean.PodDeletedMessage
				return nodeStatus, true
			}
			return nodeStatus, false
		}
		if pod.DeletionTimestamp != nil {
			nodeStatus.Message = informerBean.PodDeletedMessage
			return nodeStatus, true
		}
	}
	return nodeStatus, false
}

func (impl *InformerImpl) assessNodeStatus(eventType string, pod *coreV1.Pod) v1alpha1.NodeStatus {
	nodeStatus := v1alpha1.NodeStatus{}

	/*
		Special handling for delete events with force delete scenarios, Kubernetes does NOT guarantee that the pod status
		will be updated to "Failed" during force delete. Sometimes the pod phase can be "Running" even after force delete.
		Force deletion immediately removes the Pod object from the Kubernetes API server without waiting for
		the kubelet on the node to confirm termination.

		If the application within the pod on the node is still running and hasn't received a termination signal or processed
		it yet, the container processes might continue to exist on the node even after the Pod object is gone from the API
		server. This can lead to a state where the pod effectively exists on the node, but Kubernetes no longer tracks it,
		and it might appear as Running if you were to inspect the node's process list.
	*/
	if eventType == bean.DeleteEvent && isPodForceDeletedWhileRunning(pod) {
		// Force delete detected - treat as failed regardless of current phase
		impl.logger.Infow("Force delete detected for pod", "podName", pod.Name, "currentPhase", pod.Status.Phase, "deletionGracePeriod", *pod.DeletionGracePeriodSeconds)
		nodeStatus.Phase = v1alpha1.NodeFailed
		nodeStatus.Message = bean.NodeForceDeleted
		return nodeStatus
	} else {
		switch pod.Status.Phase {
		case coreV1.PodPending:
			nodeStatus.Phase = v1alpha1.NodePending
			nodeStatus.Message = getPendingReason(pod)
		case coreV1.PodSucceeded:
			nodeStatus.Phase = v1alpha1.NodeSucceeded
		case coreV1.PodFailed:
			nodeStatus.Phase, nodeStatus.Message = impl.inferFailedReason(eventType, pod)
			impl.logger.Infof("Pod %s failed: %s", pod.Name, nodeStatus.Message)
		case coreV1.PodRunning:
			nodeStatus.Phase = v1alpha1.NodeRunning
		default:
			nodeStatus.Phase = v1alpha1.NodeError
			nodeStatus.Message = fmt.Sprintf("Unexpected pod phase for %s: %s", pod.ObjectMeta.Name, pod.Status.Phase)
		}
	}

	// only update Pod IP for daemoned nodes to reduce number of updates
	if !nodeStatus.Completed() && nodeStatus.IsDaemoned() {
		nodeStatus.PodIP = pod.Status.PodIP
	}
	nodeStatus.HostNodeName = pod.Spec.NodeName

	if !nodeStatus.Progress.IsValid() {
		nodeStatus.Progress = v1alpha1.ProgressDefault
	}

	if x, ok := pod.Annotations[common.AnnotationKeyProgress]; ok {
		impl.logger.Warn("workflow uses legacy/insecure pod patch, see https://argoproj.github.io/argo-workflows/workflow-rbac/")
		if p, ok := v1alpha1.ParseProgress(x); ok {
			nodeStatus.Progress = p
		}
	}

	// We capture the exit-code after we look for the task-result.
	// All other outputs are set by the executor, only the exit-code is set by the controller.
	// By waiting, we avoid breaking the race-condition check.
	if exitCode := getExitCode(pod); exitCode != nil {
		if nodeStatus.Outputs == nil {
			nodeStatus.Outputs = &v1alpha1.Outputs{}
		}
		nodeStatus.Outputs.ExitCode = pointer.StringPtr(fmt.Sprint(*exitCode))
	}

	if nodeStatus.Fulfilled() && nodeStatus.FinishedAt.IsZero() {
		nodeStatus.FinishedAt = getLatestFinishedAt(pod)
		//nodeStatus.ResourcesDuration = durationForPod(pod)
	}

	return nodeStatus
}

func (impl *InformerImpl) inferFailedReason(eventType string, pod *coreV1.Pod) (v1alpha1.NodePhase, string) {
	if pod.Status.Message != "" {
		// Pod has a nice error message. Use that.
		return v1alpha1.NodeFailed, pod.Status.Message
	}

	// We only get one message to set for the overall node status.
	// If multiple containers failed, in order of preference:
	// init containers (will be appended later), main (annotated), main (exit code), wait, sidecars.
	order := func(n string) int {
		switch {
		case n == common.MainContainerName:
			return 1
		case n == common.WaitContainerName:
			return 2
		default:
			return 3
		}
	}
	ctrs := pod.Status.ContainerStatuses
	sort.Slice(ctrs, func(i, j int) bool { return order(ctrs[i].Name) < order(ctrs[j].Name) })
	// Init containers have the highest preferences over other containers.
	ctrs = append(pod.Status.InitContainerStatuses, ctrs...)

	for _, ctr := range ctrs {
		// In our case, we are not using init/wait containers. So ideally, there should be only main container.

		// Virtual Kubelet environment will not set the terminate on waiting container
		// https://github.com/argoproj/argo-workflows/issues/3879
		// https://github.com/virtual-kubelet/virtual-kubelet/blob/7f2a02291530d2df14905702e6d51500dd57640a/node/sync.go#L195-L208

		if ctr.State.Waiting != nil {
			return v1alpha1.NodeError, fmt.Sprintf("Pod failed before %s container starts due to %s: %s", ctr.Name, ctr.State.Waiting.Reason, ctr.State.Waiting.Message)
		}
		t := ctr.State.Terminated
		if t == nil {
			// Note: We should never get here
			// If we do, it means the pod phase is 'Failed' but the main container state is not in 'terminated' state,

			// there is a known issue.
			// when the spot node gets terminated, there can be 2 possible scenarios.
			// case1: we get the last[n] event from pod informer with the pod phase as failed and the main container state as running.
			// case2: we get the above event[n-1] and last[n] event with pod phase as failed and the main container state as terminated.

			// we want to handle the below case in last[n] event only,last event is always caught by DELETE_EVENT informer.
			if eventType == bean.DeleteEvent {
				// we should mark this case as an error
				if ctr.Name == common.MainContainerName {
					return v1alpha1.NodeFailed, impl.getFailedReasonFromPodConditions(pod.Status.Conditions)
				}
			}
			impl.logger.Warnw("Pod phase was Failed but container did not have terminated state", "podName", pod.Name, "containerName", ctr.Name, "status", pod.Status)
			continue
		}

		if t.ExitCode == 0 {
			continue
		}

		msg := fmt.Sprintf("%s (exit code %d)", t.Reason, t.ExitCode)
		if t.Message != "" {
			msg = fmt.Sprintf("%s: %s", msg, t.Message)
		}

		switch {
		case ctr.Name == common.InitContainerName:
			return v1alpha1.NodeError, msg
		case ctr.Name == common.MainContainerName:
			return v1alpha1.NodeFailed, msg
		case ctr.Name == common.WaitContainerName:
			return v1alpha1.NodeError, msg
		default:
			if t.ExitCode == 137 || t.ExitCode == 143 {
				// if the sidecar was SIGKILL'd (exit code 137) assume it was because argoexec
				// forcibly killed the container, which we ignore the error for.
				// Java code 143 is a normal exit 128 + 15 https://github.com/elastic/elasticsearch/issues/31847
				impl.logger.Infof("Ignoring %d exit code of container '%s'", t.ExitCode, ctr.Name)
			} else {
				return v1alpha1.NodeFailed, msg
			}
		}
	}

	// If we get here, we have detected that the main/wait containers succeed but the sidecar(s)
	// were  SIGKILL'd. The executor may have had to forcefully terminate the sidecar (kill -9),
	// resulting in a 137 exit code (which we had ignored earlier). If failMessages is empty, it
	// indicates that this is the case and we return Success instead of Failure.

	// Update diff from argo workflow here as we only have one main container.
	// Handling this for case of spot interruption where containers are in running state (no termination state found), in that case
	// it was marking it successful, doing this as it will be skipped at upper level, and delete event will handle it.
	// ticket - you can find debug logs/details here - https://github.com/devtron-labs/sprint-tasks/issues/2092
	impl.logger.Infow("Pod phase was Failed but no container had terminated state, marking it as failed now", "podName", pod.Name, "status", pod.Status)

	// Here we're intentionally returning empty string as message,
	// as we don't want this event to be intercepted as actual update event.
	// This case will be handled by the delete event.
	return v1alpha1.NodeFailed, ""
}

func (impl *InformerImpl) getFailedReasonFromPodConditions(conditions []coreV1.PodCondition) string {
	if len(conditions) == 0 {
		// This should never happen.
		impl.logger.Warnw("Pod phase was Failed but no conditions found")
		// If we get here, that means the pod is deleted after the container is started.
		// The Only possible reason is NodeNoLongerExists.
		return bean.NodeNoLongerExists
	}
	// Sort the conditions by last transition time, newest first.
	// This is to ensure that we get the latest reason for the failure.
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].LastTransitionTime.Time.After(conditions[j].LastTransitionTime.Time)
	})
	// If the latest condition has a message, use that.
	if conditions[0].Message != "" {
		return conditions[0].Message
	}
	// If we get here, that means the pod is deleted after the container is started.
	// The Only possible reason is NodeNoLongerExists.
	impl.logger.Warnw("Pod phase was Failed but the lastest condition has no message", "conditions", conditions)
	return bean.NodeNoLongerExists
}

// foundAnyUpdateInPodStatus return true if any of the pod status fields have changed or if the pod is new
func foundAnyUpdateInPodStatus(from *coreV1.Pod, to *coreV1.Pod) bool {
	// always expect on of the below to be not nil
	if from == nil || to == nil {
		return true
	}
	return isAnyChangeInPodStatus(&from.Status, &to.Status)
}

func isAnyChangeInPodStatus(from *coreV1.PodStatus, to *coreV1.PodStatus) bool {
	return from.Phase != to.Phase ||
		from.Message != to.Message ||
		from.PodIP != to.PodIP ||
		isAnyChangeInContainersStatus(from.ContainerStatuses, to.ContainerStatuses) ||
		isAnyChangeInContainersStatus(from.InitContainerStatuses, to.InitContainerStatuses) ||
		isAnyChangeInPodConditions(from.Conditions, to.Conditions)
}

func isAnyChangeInContainersStatus(from []coreV1.ContainerStatus, to []coreV1.ContainerStatus) bool {
	if len(from) != len(to) {
		return true
	}
	statuses := map[string]coreV1.ContainerStatus{}
	for _, s := range from {
		statuses[s.Name] = s
	}
	for _, s := range to {
		if isAnyChangeInContainerStatus(statuses[s.Name], s) {
			return true
		}
	}
	return false
}

func isAnyChangeInContainerStatus(from coreV1.ContainerStatus, to coreV1.ContainerStatus) bool {
	return from.Ready != to.Ready || isAnyChangeInContainerState(from.State, to.State)
}

func isAnyChangeInContainerState(from coreV1.ContainerState, to coreV1.ContainerState) bool {
	// waiting has two significant fields and either could potentially change
	return to.Waiting != nil && (from.Waiting == nil || from.Waiting.Message != to.Waiting.Message || from.Waiting.Reason != to.Waiting.Reason) ||
		// running only has one field which is immutable -  so any change is significant
		(to.Running != nil && from.Running == nil) ||
		// I'm assuming this field is immutable - so any change is significant
		(to.Terminated != nil && from.Terminated == nil)
}

func isAnyChangeInPodConditions(from []coreV1.PodCondition, to []coreV1.PodCondition) bool {
	if len(from) != len(to) {
		return true
	}
	for i, a := range from {
		b := to[i]
		if a.Message != b.Message || a.Reason != b.Reason {
			return true
		}
	}
	return false
}
