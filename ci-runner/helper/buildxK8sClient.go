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
	"log"
	"net/http"
	"time"
)

type BuildxK8sInterface interface {
	PatchOwnerReferenceInBuilders()
	RegisterBuilderPods(ctx context.Context) error
	CatchBuilderPodLivenessError(ctx context.Context) error
	WaitUntilBuilderPodLive(ctx context.Context, done chan<- bool)
}

type buildxK8sClient struct {
	restConfig      *rest.Config
	httpClient      *http.Client
	appV1Client     *appsV1.AppsV1Client
	coreV1Client    *clientcorev1.CoreV1Client
	namespace       string
	deploymentNames []string
	podNames        []string
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
	return &buildxK8sClient{
		restConfig:      restConfig,
		httpClient:      k8sHttpClient,
		appV1Client:     appV1ClientSet,
		coreV1Client:    coreV1ClientSet,
		namespace:       "devtron-ci",
		deploymentNames: deploymentNames,
	}, nil
}

func (k8s *buildxK8sClient) PatchOwnerReferenceInBuilders() {
	if k8s == nil {
		return
	}
	for _, deploymentName := range k8s.deploymentNames {
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
		err := k8s.builderPodLivenessDialer(ctx)
		if err != nil {
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
		if err := k8s.builderPodLivenessDialer(ctx); err == nil {
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
	k8s.podNames = make([]string, 0)
	for _, deploymentName := range k8s.deploymentNames {
		podNames, err := k8s.getBuilderPods(ctx, deploymentName)
		if err != nil {
			log.Println(util.DEVTRON, fmt.Sprintf("error while getting builder pods for deployment: %q, err: %v", deploymentName, err))
			return err
		}
		if len(podNames) == 0 {
			log.Println(util.DEVTRON, fmt.Sprintf("no running builder pods found for deployment: %q", deploymentName))
			return BuilderPodDeletedError
		}
		k8s.podNames = append(k8s.podNames, podNames...)
		log.Println(util.DEVTRON, fmt.Sprintf("registered builder pods for deployment: %q, pods: %v", deploymentName, podNames))
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

func (k8s *buildxK8sClient) builderPodLivenessDialer(ctx context.Context) error {
	for _, podName := range k8s.podNames {
		isRunning, err := k8s.verifyRunningBuilder(ctx, podName)
		if err != nil && !k8sError.IsNotFound(err) {
			log.Println(util.DEVTRON, fmt.Sprintf("error while verifying running builders for pod: %q, err: %v", podName, err))
			return err
		} else if k8sError.IsNotFound(err) || !isRunning {
			log.Println(util.DEVTRON, fmt.Sprintf("builder pod liveness failed for pod: %q", podName))
			return BuilderPodDeletedError
		}
	}
	log.Println(util.DEVTRON, "builder pod liveness check passed")
	return nil
}

func (k8s *buildxK8sClient) verifyRunningBuilder(ctx context.Context, podName string) (bool, error) {
	pod, err := k8s.coreV1Client.Pods("devtron-ci").Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if pod.Status.Phase == corev1.PodRunning {
		log.Println(util.DEVTRON, fmt.Sprintf("Pod %q is running", pod.Name))
		return true, err
	}
	log.Println(util.DEVTRON, fmt.Sprintf("Pod %q is not running, current phase: %q", pod.Name, pod.Status.Phase))
	return false, err
}

func (k8s *buildxK8sClient) getBuilderPods(ctx context.Context, deploymentName string) ([]string, error) {
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
	podNames := make([]string, 0, len(podList.Items))
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			log.Println(util.DEVTRON, fmt.Sprintf("Pod %q is running", pod.Name))
			podNames = append(podNames, pod.Name)
		}
	}
	return podNames, err
}
