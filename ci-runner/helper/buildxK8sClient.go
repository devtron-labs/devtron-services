package helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/ci-runner/util"
	"github.com/devtron-labs/common-lib/informer"
	"github.com/devtron-labs/common-lib/utils"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	appsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"net/http"
	"time"
)

type BuildxK8sInterface interface {
	PatchOwnerReferenceInBuilders(deploymentNames []string)
	BuilderPodLivenessDialer(ctx context.Context, deploymentNames []string) error
}

type buildxK8sClient struct {
	restConfig   *rest.Config
	httpClient   *http.Client
	appV1Client  *appsV1.AppsV1Client
	coreV1Client *clientcorev1.CoreV1Client
	namespace    string
}

func newBuildxK8sClient() (*buildxK8sClient, error) {
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
		restConfig:   restConfig,
		httpClient:   k8sHttpClient,
		appV1Client:  appV1ClientSet,
		coreV1Client: coreV1ClientSet,
		namespace:    "devtron-ci",
	}, nil
}

func (k8s *buildxK8sClient) PatchOwnerReferenceInBuilders(deploymentNames []string) {
	for _, deploymentName := range deploymentNames {
		if err := k8s.jsonPatchOwnerReferenceInDeployment(deploymentName); err != nil {
			fmt.Println(util.DEVTRON, "failed to patch the buildkit deployment's owner reference, ", " deployment: ", deploymentName, " err: ", err)
		} else {
			fmt.Println(util.DEVTRON, "successfully patched the buildkit deployment's owner reference, ", " deployment: ", deploymentName)
		}
	}
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

func (k8s *buildxK8sClient) builderPodLivenessDialer(ctx context.Context, deploymentNames []string) error {
	for _, deploymentName := range deploymentNames {
		isRunning, err := k8s.verifyRunningBuilders(ctx, deploymentName)
		if err != nil && !k8sError.IsNotFound(err) {
			fmt.Println(util.DEVTRON, fmt.Sprintf("error while verifying running builders for deployment: %q, err: %v", deploymentName, err))
			return err
		} else if k8sError.IsNotFound(err) || !isRunning {
			fmt.Println(util.DEVTRON, fmt.Sprintf("builder pod liveness failed for deployment: %q", deploymentName))
			return errors.New(informer.PodDeletedMessage)
		}
	}
	fmt.Println(util.DEVTRON, "builder pod liveness check passed for all deployments")
	return nil
}

func (k8s *buildxK8sClient) BuilderPodLivenessDialer(ctx context.Context, deploymentNames []string) error {
	for {
		err := k8s.builderPodLivenessDialer(ctx, deploymentNames)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			fmt.Println(util.DEVTRON, "context done, exiting builder pod liveness check")
			return nil
		default:
			fmt.Println(util.DEVTRON, "sleeping for 10 seconds before next builder pod liveness check")
			// Sleep for 10 seconds
			<-time.After(10 * time.Second)
		}
	}
}

func (k8s *buildxK8sClient) verifyRunningBuilders(ctx context.Context, deploymentName string) (bool, error) {
	deploy, err := k8s.appV1Client.Deployments("devtron-ci").Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return false, err
	}
	listOpts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	podList, err := k8s.coreV1Client.Pods(k8s.namespace).List(ctx, listOpts)
	if err != nil {
		return false, err
	}
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			fmt.Println(util.DEVTRON, fmt.Sprintf("Pod %q is running", pod.Name))
			return true, err
		} else {
			fmt.Println(util.DEVTRON, fmt.Sprintf("Pod %q is not running, current phase: %q", pod.Name, pod.Status.Phase))
			return false, err
		}
	}
	return false, err
}
