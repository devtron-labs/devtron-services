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

package multiCluster

import (
	"encoding/json"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	repository "github.com/devtron-labs/kubewatch/pkg/cluster"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/devtron-labs/kubewatch/pkg/middleware"
	"github.com/devtron-labs/kubewatch/pkg/resource"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	"github.com/devtron-labs/kubewatch/pkg/utils"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"log"
	"strconv"
	"time"
)

type Informer interface {
	Start() error
	Stop()
}

type InformerImpl struct {
	logger                *zap.SugaredLogger
	clusterRepository     repository.ClusterRepository
	defaultK8sConfig      *rest.Config
	pubSubClient          *pubsub.PubSubClientServiceImpl
	appConfig             *config.AppConfig
	k8sUtil               utils.K8sUtil
	informerClient        resource.InformerClient
	defaultClusterStopper *informerFactoryStopper
	multiClusterStopper   map[int]*sharedInformerStopper
	// TODO Asutosh: should we introduce lock??
}

func NewMultiClusterInformerImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository,
	client *pubsub.PubSubClientServiceImpl,
	appConfig *config.AppConfig,
	k8sUtil utils.K8sUtil,
	informerClient resource.InformerClient,
	defaultK8sConfig *rest.Config) (*InformerImpl, error) {
	if !appConfig.IsMultiClusterMode() {
		logger.Warn("multi cluster mode is disabled")
		logger.Debugw("multi cluster mode is disabled", "appConfig", appConfig)
		return nil, nil
	}
	if clusterRepository == nil {
		return nil, errors.New("clusterRepository is required for Informer")
	}
	multiClusterStopper := make(map[int]*sharedInformerStopper)
	return &InformerImpl{
		logger:              logger,
		clusterRepository:   clusterRepository,
		pubSubClient:        client,
		appConfig:           appConfig,
		k8sUtil:             k8sUtil,
		informerClient:      informerClient,
		defaultK8sConfig:    defaultK8sConfig,
		multiClusterStopper: multiClusterStopper,
	}, nil
}

func (impl *InformerImpl) Start() error {
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start informer")
	err := impl.startDefaultClusterInformer()
	if err != nil {
		impl.logger.Errorw("error in starting default cluster informer", "err", err)
		middleware.IncUnregisteredInformers("Default cluster", "1", "DefaultClusterSecret")
		return err
	}
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		impl.logger.Error("error in fetching clusters", "err", err)
		return err
	}
	for _, model := range models {
		if err = impl.startInformer(model.Id); err != nil {
			impl.logger.Errorw("error in starting informer for cluster", "clusterId", model.Id, "err", err)
			// ignore error and continue with other clusters
		}
	}
	return nil
}

func (impl *InformerImpl) Stop() {
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start default cluster informer")
	for _, clusterId := range maps.Keys(impl.multiClusterStopper) {
		impl.stopInformerByClusterId(clusterId)
	}
	impl.stopDefaultClusterInformer()
}

func (impl *InformerImpl) stopDefaultClusterInformer() {
	if impl.defaultClusterStopper != nil {
		impl.defaultClusterStopper.Stop()
	}
}

func (impl *InformerImpl) stopInformerByClusterId(clusterId int) {
	if stopper, ok := impl.multiClusterStopper[clusterId]; ok {
		stopper.Stop()
		delete(impl.multiClusterStopper, clusterId)
	}
}

// cluster informer starts ------

func (impl *InformerImpl) startDefaultClusterInformer() error {
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start default cluster informer")
	impl.logger.Debug("starting informer, reading new cluster request for default cluster")
	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.FieldSelector = CLUSTER_MODIFY_EVENT_FIELD_SELECTOR
	})
	// addFunc is called when a new secret is created
	addFunc := func(secretObject *coreV1.Secret) {
		if err := impl.handleClusterChangeEvent(secretObject); err != nil {
			impl.logger.Errorw("error in handling cluster add event", "err", err, "secretObject", secretObject)
		}
	}
	// updateFunc is called when an existing secret is updated
	updateFunc := func(oldSecretObject, newSecretObject *coreV1.Secret) {
		if err := impl.handleClusterChangeEvent(newSecretObject); err != nil {
			impl.logger.Errorw("error in handling cluster update event", "err", err, "newSecretObject", newSecretObject)
		}
	}
	// deleteFunc is called when an existing secret is deleted
	deleteFunc := func(secretObject *coreV1.Secret) {
		if err := impl.handleClusterDeleteEvent(secretObject); err != nil {
			impl.logger.Errorw("error in handling cluster delete event", "err", err, "secretObject", secretObject)
		}
	}
	informerFactory := impl.informerClient.GetSecretInformerFactory()
	eventHandler := bean.NewEventHandlers[coreV1.Secret]().
		AddFuncHandler(addFunc).
		UpdateFuncHandler(updateFunc).
		DeleteFuncHandler(deleteFunc)
	secretFactory, err := informerFactory.GetSharedInformerFactory(impl.defaultK8sConfig, eventHandler, labelOptions)
	if err != nil {
		impl.logger.Errorw("error in registering default cluster secret informer", "err", err)
		return err
	}
	stopChannel := make(chan struct{})
	impl.defaultClusterStopper = newInformerFactoryStopper(secretFactory, stopChannel)
	secretFactory.Start(stopChannel)
	// waiting for cache sync
	synced := secretFactory.WaitForCacheSync(stopChannel)
	for v, ok := range synced {
		if !ok {
			impl.logger.Errorw("failed to sync secret informer for default cluster", "value", v)
			return errors.New("failed to sync secret informer for default cluster")
		}
	}
	return nil
}

func (impl *InformerImpl) handleClusterChangeEvent(secretObject *coreV1.Secret) error {
	if secretObject.Type != CLUSTER_MODIFY_EVENT_SECRET_TYPE {
		return nil
	}
	data := secretObject.Data
	action := data[SECRET_FIELD_ACTION]
	id := string(data[SECRET_FIELD_CLUSTER_ID])
	clusterId, convErr := strconv.Atoi(id)
	if convErr != nil {
		impl.logger.Errorw("error in converting cluster id to int", "clusterId", id, "err", convErr)
		return convErr
	}
	if string(action) == CLUSTER_ACTION_ADD {
		if err := impl.startInformer(clusterId); err != nil {
			impl.logger.Errorw("error in starting informer for cluster", "clusterId", clusterId, "err", err)
			return err
		}
	} else if string(action) == CLUSTER_ACTION_UPDATE {
		if err := impl.syncMultiClusterInformer(clusterId); err != nil {
			impl.logger.Errorw("error in updating informer for cluster", "id", clusterId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *InformerImpl) handleClusterDeleteEvent(secretObject *coreV1.Secret) error {
	if secretObject.Type != CLUSTER_MODIFY_EVENT_SECRET_TYPE {
		return nil
	}
	data := secretObject.Data
	action := data[SECRET_FIELD_ACTION]
	id := string(data[SECRET_FIELD_CLUSTER_ID])
	clusterId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if string(action) == CLUSTER_ACTION_DELETE {
		if err = impl.handleClusterDelete(clusterId); err != nil {
			impl.logger.Errorw("error in handling cluster delete event", "clusterId", clusterId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *InformerImpl) handleClusterDelete(clusterId int) error {
	deleteClusterInfo, err := impl.clusterRepository.FindByIdWithActiveFalse(clusterId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching cluster by id", "cluster-id ", clusterId, "err", err)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("cluster not found", "clusterId", clusterId)
		return nil
	}
	impl.stopInformerByClusterId(deleteClusterInfo.Id)
	return nil
}

// cluster informer ends ------

// sync informer starts ------

func (impl *InformerImpl) syncMultiClusterInformer(clusterId int) error {
	clusterInfo, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Error("error in fetching cluster info by id", "err", err)
		return err
	}
	// before creating a new informer for cluster, close the existing one
	impl.logger.Debugw("stopping informer for cluster", "cluster-name", clusterInfo.ClusterName, "cluster-id", clusterInfo.Id)
	impl.stopInformerByClusterId(clusterInfo.Id)
	impl.logger.Debugw("informer stopped", "cluster-name", clusterInfo.ClusterName, "cluster-id", clusterInfo.Id)
	// create new informer for cluster with new config
	err = impl.startInformer(clusterInfo.Id)
	if err != nil {
		impl.logger.Errorw("error in starting informer for cluster", "clusterId", clusterInfo.Id, "err", err)
		return err
	}
	return nil
}

// sync informer ends ------

// register informer starts ------

func (impl *InformerImpl) startInformer(clusterId int) error {
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start informer", "clusterId", clusterId)
	clusterInfo, err := impl.clusterRepository.FindById(clusterId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching cluster", "clusterId", clusterId, "err", err)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("cluster not found", "clusterId", clusterId)
		return nil
	}
	if len(clusterInfo.ErrorInConnecting) > 0 {
		impl.logger.Debugw("cluster is not reachable", "clusterId", clusterId, "clusterName", clusterInfo.ClusterName)
		middleware.IncUnreachableCluster(clusterInfo.ClusterName, strconv.Itoa(clusterInfo.Id))
	}
	if impl.appConfig.IsMultiClusterSystemExec() {
		err = impl.startSystemExecInformer(clusterInfo)
		if err != nil && !errors.Is(err, AlreadyExists) {
			impl.logger.Errorw("error in starting system executor informer for cluster", "clusterId", clusterId, "err", err)
			middleware.IncUnregisteredInformers(clusterInfo.ClusterName, strconv.Itoa(clusterInfo.Id), "SystemExecutor")
			return err
		} else if errors.Is(err, AlreadyExists) {
			impl.logger.Warnw("system executor informer already exist for cluster", "clusterId", clusterId)
		}
	}
	if impl.appConfig.IsMultiClusterArgoCD() {
		err = impl.startArgoCdInformer(clusterInfo)
		if err != nil && !errors.Is(err, AlreadyExists) {
			impl.logger.Errorw("error in starting argo cd informer for cluster", "clusterId", clusterId, "err", err)
			middleware.IncUnregisteredInformers(clusterInfo.ClusterName, strconv.Itoa(clusterInfo.Id), "ArgoCD")
			return err
		} else if errors.Is(err, AlreadyExists) {
			impl.logger.Warnw("argo cd informer already exist for cluster", "clusterId", clusterId)
		}
	}
	return nil
}

func (impl *InformerImpl) startSystemExecInformer(clusterInfo *repository.Cluster) error {
	if !clusterInfo.SupportsSystemExec() {
		impl.logger.Warnw("argo workflow setup is not done for cluster, skipping", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
		return nil
	}
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start system workflow informer", "clusterId", clusterInfo.Id)
	stopper, ok := impl.multiClusterStopper[clusterInfo.Id]
	if ok && stopper.HasSystemWfInformer() {
		impl.logger.Debug(fmt.Sprintf("informer for %s already exist", clusterInfo.ClusterName))
		return AlreadyExists
	}
	impl.logger.Infow("starting informer for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	restConfig := impl.getK8sConfigForCluster(clusterInfo)
	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = WORKFLOW_LABEL_SELECTOR
	})
	// updateFunc is called when an existing pod is updated
	updateFunc := func(oldPodObj, newPodObj *coreV1.Pod) {
		var workflowType string
		if newPodObj.Labels != nil {
			if val, ok := newPodObj.Labels[WORKFLOW_TYPE_LABEL_KEY]; ok {
				workflowType = val
			}
		}
		impl.logger.Debugw("event received in Pods update informer", "time", time.Now(), "podObjStatus", newPodObj.Status)
		nodeStatus := impl.assessNodeStatus(newPodObj)
		workflowStatus := getWorkflowStatus(newPodObj, nodeStatus, workflowType)
		wfJson, err := json.Marshal(workflowStatus)
		if err != nil {
			impl.logger.Errorw("error occurred while marshalling workflowJson", "err", err)
			return
		}
		impl.logger.Debugw("sending system executor workflow update event", "workflow", string(wfJson))
		if impl.pubSubClient == nil {
			log.Println("don't publish")
			return
		}
		topic, err := getTopic(workflowType)
		if err != nil {
			impl.logger.Errorw("error while getting Topic")
			return
		}
		err = impl.pubSubClient.Publish(topic, string(wfJson))
		if err != nil {
			impl.logger.Errorw("error while publishing Request", "err", err)
			return
		}

		impl.logger.Debug("cd workflow update sent")
	}
	// deleteFunc is called when an existing pod is deleted
	deleteFunc := func(podObj *coreV1.Pod) {
		var workflowType string
		if podObj.Labels != nil {
			if val, ok := podObj.Labels[WORKFLOW_TYPE_LABEL_KEY]; ok {
				workflowType = val
			}
		}
		impl.logger.Debugw("event received in Pods delete informer", "time", time.Now(), "podObjStatus", podObj.Status)
		nodeStatus := impl.assessNodeStatus(podObj)
		nodeStatus, reTriggerRequired := impl.checkIfPodDeletedAndUpdateMessage(podObj.Name, podObj.Namespace, nodeStatus, restConfig)
		if !reTriggerRequired {
			//not sending this deleted event if it's not a re-trigger case
			return
		}
		workflowStatus := getWorkflowStatus(podObj, nodeStatus, workflowType)
		wfJson, err := json.Marshal(workflowStatus)
		if err != nil {
			impl.logger.Errorw("error occurred while marshalling workflowJson", "err", err)
			return
		}
		impl.logger.Debugw("sending system executor cd workflow delete event", "workflow", string(wfJson))
		if impl.pubSubClient == nil {
			log.Println("don't publish")
			return
		}
		topic, err := getTopic(workflowType)
		if err != nil {
			impl.logger.Errorw("error while getting Topic")
			return
		}

		err = impl.pubSubClient.Publish(topic, string(wfJson))
		if err != nil {
			impl.logger.Errorw("error while publishing Request", "err", err)
			return
		}

		impl.logger.Debug("cd workflow update sent")
	}
	podInformerFactory := impl.informerClient.GetPodInformerFactory()
	eventHandler := bean.NewEventHandlers[coreV1.Pod]().
		UpdateFuncHandler(updateFunc).
		DeleteFuncHandler(deleteFunc)
	informerFactory, err := podInformerFactory.GetSharedInformerFactory(restConfig, eventHandler, labelOptions)
	if err != nil {
		impl.logger.Errorw("error in adding event handler for cluster pod informer", "err", err)
	}
	stopChannel := make(chan struct{})
	stopper = stopper.RegisterSystemWfStopper(informerFactory, stopChannel)
	informerFactory.Start(stopChannel)
	impl.logger.Infow("informer started for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	impl.multiClusterStopper[clusterInfo.Id] = stopper
	return nil
}

func (impl *InformerImpl) startArgoCdInformer(clusterInfo *repository.Cluster) error {
	if !clusterInfo.SupportsArgoCd() {
		impl.logger.Warnw("argo cd setup is not done for cluster, skipping", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
		return nil
	}
	startTime := time.Now()
	defer utils.LogExecutionTime(impl.logger, startTime, "time taken to start argo cd informer", "clusterId", clusterInfo.Id)
	stopper, ok := impl.multiClusterStopper[clusterInfo.Id]
	if ok && stopper.HasArgoCdInformer() {
		impl.logger.Debug(fmt.Sprintf("informer for %s already exist", clusterInfo.ClusterName))
		return AlreadyExists
	}
	impl.logger.Infow("starting informer for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	cfg := impl.getK8sConfigForCluster(clusterInfo)
	acdInformer := impl.informerClient.GetSharedInformerClient(bean.ApplicationResourceType)
	informerFactory, err := acdInformer.GetSharedInformer(metav1.NamespaceAll, cfg)
	if err != nil {
		impl.logger.Errorw("error in registering acd informer", "err", err, "clusterId", clusterInfo.Id)
		return err
	}
	stopChannel := make(chan struct{})
	stopper = stopper.RegisterArgoCdStopper(stopChannel)
	informerFactory.Run(stopChannel)
	impl.logger.Infow("informer started for cluster", "clusterId", clusterInfo.Id, "clusterName", clusterInfo.ClusterName)
	impl.multiClusterStopper[clusterInfo.Id] = stopper
	return nil
}

// register informer ends ------
