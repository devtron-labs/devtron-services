package helper

import (
	"context"
	"fmt"
	"github.com/devtron-labs/ci-runner/util"
	"github.com/devtron-labs/common-lib/utils"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	appsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"log"
	"net/http"
	"slices"
	"time"
)

type BuildxK8sInterface interface {
	PatchOwnerReferenceInBuilders()
	RegisterBuilderPods(ctx context.Context) error
	RestartBuilders(ctx context.Context) error
	CatchBuilderPodLivenessError(ctx context.Context) error
	WaitUntilBuilderPodLive(ctx context.Context, done chan<- bool)
}

type buildxK8sClient struct {
	restConfig   *rest.Config
	httpClient   *http.Client
	appV1Client  *appsV1.AppsV1Client
	coreV1Client *clientcorev1.CoreV1Client
	namespace    string
	deployments  map[string][]podStatus
}

type podStatus struct {
	name  string
	phase corev1.PodPhase
	err   error
}

func newBuildxK8sClient(deploymentNames []string) (*buildxK8sClient, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sHttpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, err
	}
	appV1ClientSet, err := appsV1.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		return nil, err
	}
	coreV1ClientSet, err := clientcorev1.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		return nil, err
	}
	deployments := make(map[string][]podStatus)
	for _, deploymentName := range deploymentNames {
		deployments[deploymentName] = make([]podStatus, 0)
	}
	return &buildxK8sClient{
		restConfig:   restConfig,
		httpClient:   k8sHttpClient,
		appV1Client:  appV1ClientSet,
		coreV1Client: coreV1ClientSet,
		namespace:    "devtron-ci",
		deployments:  deployments,
	}, nil
}

func (k8s *buildxK8sClient) PatchOwnerReferenceInBuilders() {
	if k8s == nil {
		return
	}
	for deploymentName := range k8s.deployments {
		if err := k8s.jsonPatchOwnerReferenceInDeployment(deploymentName); err != nil {
			log.Println(util.DEVTRON, "failed to patch the buildkit deployment's owner reference, ", " deployment: ", deploymentName, " err: ", err)
		} else {
			log.Println(util.DEVTRON, "successfully patched the buildkit deployment's owner reference, ", " deployment: ", deploymentName)
		}
	}
}

func (k8s *buildxK8sClient) CatchBuilderPodLivenessError(ctx context.Context) error {
	if k8s == nil {
		return nil
	}
	for {
		err := k8s.builderLivenessDialer(ctx)
		if err != nil && ctx.Err() == nil {
			return err
		}
		select {
		case <-ctx.Done():
			log.Println(util.DEVTRON, "context done, exiting builder pod liveness check")
			return nil
		default:
			log.Println(util.DEVTRON, "sleeping for 10 seconds before next builder pod liveness check")
			// Sleep for 10 seconds
			<-time.After(10 * time.Second)
		}
	}
}

func (k8s *buildxK8sClient) WaitUntilBuilderPodLive(ctx context.Context, done chan<- bool) {
	if k8s == nil {
		return
	}
	for {
		if err := k8s.builderLivenessDialer(ctx); err == nil {
			done <- true
			return
		}
		select {
		case <-ctx.Done():
			log.Println(util.DEVTRON, "context done, exiting builder pod liveness check")
			return
		default:
			log.Println(util.DEVTRON, "sleeping for 10 seconds before next builder pod liveness check")
			// Sleep for 10 seconds
			<-time.After(10 * time.Second)
		}
	}
}

func (k8s *buildxK8sClient) RegisterBuilderPods(ctx context.Context) error {
	if k8s == nil {
		return nil
	}
	for deploymentName := range k8s.deployments {
		pods, err := k8s.getBuilderPods(ctx, deploymentName)
		if err != nil {
			log.Println(util.DEVTRON, fmt.Sprintf("error while getting builder pods for deployment: %q, err: %v", deploymentName, err))
			return err
		}
		k8s.deployments[deploymentName] = pods
		log.Println(util.DEVTRON, fmt.Sprintf("registered builder pods for deployment: %q", deploymentName))
	}
	return nil
}

func (k8s *buildxK8sClient) RestartBuilders(ctx context.Context) error {
	for _, pods := range k8s.deployments {
		if err := k8s.rotateBuilders(ctx, pods); err != nil {
			return err
		}
	}
	return nil
}

func (k8s *buildxK8sClient) jsonPatchOwnerReferenceInDeployment(deploymentName string) error {
	patchStr := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"apiVersion":"v1","kind":"Pod","name":"%s","uid":"%s"}]}}`, utils.GetSelfK8sPodName(), utils.GetSelfK8sUID())
	// Apply the patch directly
	// the namespace is hardcoded to devtron-ci as our k8s driver is only supported for ci's running in devtron-ci namespace.
	_, err := k8s.appV1Client.Deployments(k8s.namespace).
		Patch(
			context.TODO(),
			deploymentName,
			types.StrategicMergePatchType,
			[]byte(patchStr),
			metav1.PatchOptions{FieldManager: "patch"},
		)
	if err != nil {
		return err
	}
	return nil
}

func (k8s *buildxK8sClient) builderLivenessDialer(ctx context.Context) error {
	for deployment, pods := range k8s.deployments {
		updatedPods := k8s.getLivePodStatus(ctx, pods)
		k8s.deployments[deployment] = updatedPods
	}
	for deployment, pods := range k8s.deployments {
		if !slices.ContainsFunc(pods, func(pod podStatus) bool {
			return pod.phase == corev1.PodRunning && pod.err == nil
		}) {
			log.Println(util.DEVTRON, fmt.Sprintf("builder pod liveness check failed for deployment: %q", deployment))
			return BuilderPodDeletedError
		}
	}
	log.Println(util.DEVTRON, "builder pod liveness check passed")
	return nil
}

func (k8s *buildxK8sClient) rotateBuilders(ctx context.Context, pods []podStatus) error {
	deleteOptions := metav1.DeleteOptions{}
	deleteOptions.GracePeriodSeconds = ptr.To(int64(0))
	for _, pod := range pods {
		if pod.phase == corev1.PodRunning || pod.err == nil {
			err := k8s.coreV1Client.Pods(k8s.namespace).Delete(ctx, pod.name, deleteOptions)
			if err != nil && !k8sError.IsNotFound(err) {
				log.Println(util.DEVTRON, fmt.Sprintf("error while deleting pod: %q, err: %v", pod.name, err))
				return err
			}
		}
	}
	return nil
}

func (k8s *buildxK8sClient) getLivePodStatus(ctx context.Context, pods []podStatus) []podStatus {
	updatedPods := make([]podStatus, 0, len(pods))
	for _, pod := range pods {
		updatedPod := podStatus{
			name: pod.name,
		}
		podPhase, err := k8s.getBuilderPhase(ctx, pod.name)
		updatedPod.phase = podPhase
		if err != nil && !k8sError.IsNotFound(err) {
			log.Println(util.DEVTRON, fmt.Sprintf("error while verifying running builders for pod: %q, err: %v", pod.name, err))
			updatedPod.err = err
		} else if k8sError.IsNotFound(err) || podPhase != corev1.PodRunning {
			log.Println(util.DEVTRON, fmt.Sprintf("builder pod liveness failed for pod: %q, phase: %q", pod.name, podPhase))
			updatedPod.err = BuilderPodDeletedError
		}
		updatedPods = append(updatedPods, updatedPod)
	}
	return updatedPods
}

func (k8s *buildxK8sClient) getBuilderPhase(ctx context.Context, podName string) (corev1.PodPhase, error) {
	pod, err := k8s.coreV1Client.Pods("devtron-ci").Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return corev1.PodUnknown, err
	}
	return pod.Status.Phase, err
}

func (k8s *buildxK8sClient) getBuilderPods(ctx context.Context, deploymentName string) ([]podStatus, error) {
	deploy, err := k8s.appV1Client.Deployments("devtron-ci").Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return nil, err
	}
	listOpts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	podList, err := k8s.coreV1Client.Pods(k8s.namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	podNames := make([]podStatus, 0, len(podList.Items))
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed ||
			pod.DeletionTimestamp != nil {
			log.Println(util.DEVTRON, fmt.Sprintf("ignoring pod %q as it is not in running state, phase: %q", pod.Name, pod.Status.Phase))
			continue
		}
		// register the running pod only
		podNames = append(podNames, podStatus{
			name:  pod.Name,
			phase: pod.Status.Phase,
		})
	}
	return podNames, err
}
