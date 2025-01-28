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
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/informer/errors"
	"golang.org/x/exp/maps"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sort"
)

func (impl *InformerImpl) getSystemWfStopper(clusterId int) (*informerBean.FactoryStopper, bool) {
	stopper, ok := impl.systemWfInformerStopper[clusterId]
	if ok {
		return stopper, stopper.HasInformer()
	}
	return stopper, false
}

func (impl *InformerImpl) getStoppableClusterIds() []int {
	return maps.Keys(impl.systemWfInformerStopper)
}

func (impl *InformerImpl) getStopChannel(informerFactory kubeinformers.SharedInformerFactory, clusterLabels *informerBean.ClusterLabels) (chan struct{}, error) {
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
	if (nodeStatus.Phase == v1alpha1.NodeFailed || nodeStatus.Phase == v1alpha1.NodeError) && nodeStatus.Message == informerBean.EXIT_CODE_143_ERROR {
		clusterClient, k8sErr := impl.k8sUtil.GetK8sClientForConfig(config)
		if k8sErr != nil {
			return nodeStatus, false
		}
		pod, err := clusterClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting pod from clusterClient", "podName", podName, "namespace", namespace, "err", err)
			if isResourceNotFoundErr(err) {
				nodeStatus.Message = informerBean.POD_DELETED_MESSAGE
				return nodeStatus, true
			}
			return nodeStatus, false
		}
		if pod.DeletionTimestamp != nil {
			nodeStatus.Message = informerBean.POD_DELETED_MESSAGE
			return nodeStatus, true
		}
	}
	return nodeStatus, false
}

func (impl *InformerImpl) assessNodeStatus(pod *coreV1.Pod) v1alpha1.NodeStatus {
	nodeStatus := v1alpha1.NodeStatus{}
	switch pod.Status.Phase {
	case coreV1.PodPending:
		nodeStatus.Phase = v1alpha1.NodePending
		nodeStatus.Message = getPendingReason(pod)
	case coreV1.PodSucceeded:
		nodeStatus.Phase = v1alpha1.NodeSucceeded
	case coreV1.PodFailed:
		nodeStatus.Phase, nodeStatus.Message = impl.inferFailedReason(pod)
		impl.logger.Infof("Pod %s failed: %s", pod.Name, nodeStatus.Message)
	case coreV1.PodRunning:
		nodeStatus.Phase = v1alpha1.NodeRunning
	default:
		nodeStatus.Phase = v1alpha1.NodeError
		nodeStatus.Message = fmt.Sprintf("Unexpected pod phase for %s: %s", pod.ObjectMeta.Name, pod.Status.Phase)
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

func (impl *InformerImpl) inferFailedReason(pod *coreV1.Pod) (v1alpha1.NodePhase, string) {
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

		// Virtual Kubelet environment will not set the terminate on waiting container
		// https://github.com/argoproj/argo-workflows/issues/3879
		// https://github.com/virtual-kubelet/virtual-kubelet/blob/7f2a02291530d2df14905702e6d51500dd57640a/node/sync.go#L195-L208

		if ctr.State.Waiting != nil {
			return v1alpha1.NodeError, fmt.Sprintf("Pod failed before %s container starts due to %s: %s", ctr.Name, ctr.State.Waiting.Reason, ctr.State.Waiting.Message)
		}
		t := ctr.State.Terminated
		if t == nil {
			// We should never get here
			impl.logger.Warnf("Pod %s phase was Failed but %s did not have terminated state", pod.Name, ctr.Name)
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
	return v1alpha1.NodeSucceeded, ""
}
