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

package k8sInformer

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/async"
	k8sUtils "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/kubelink/bean"
	globalConfig "github.com/devtron-labs/kubelink/config"
	"github.com/devtron-labs/kubelink/converter"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/internals/middleware"
	repository "github.com/devtron-labs/kubelink/pkg/cluster"
	"github.com/devtron-labs/kubelink/pkg/service/helmApplicationService/adapter"
	"go.uber.org/zap"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const (
	HELM_RELEASE_SECRET_TYPE         = "helm.sh/release.v1"
	CLUSTER_MODIFY_EVENT_SECRET_TYPE = "cluster.request/modify"
	DEFAULT_CLUSTER                  = "default_cluster"
	INFORMER_ALREADY_EXIST_MESSAGE   = "INFORMER_ALREADY_EXIST"
	ADD                              = "add"
	UPDATE                           = "update"
	DELETE                           = "delete"
)

type K8sInformer interface {
	GetAllReleaseByClusterId(clusterId int) []*client.DeployedAppDetail
	CheckReleaseExists(clusterId int32, releaseIdentifier string) bool
	GetClusterClientSet(clusterInfo bean.ClusterInfo) (*kubernetes.Clientset, error)
	RegisterListener(listener ClusterSecretUpdateListener)
	GetReleaseDetails(clusterId int32, releaseIdentifier string) (*client.DeployedAppDetail, error)
}

type ClusterSecretUpdateListener interface {
	OnStateChange(clusterId int, action string)
}

type K8sInformerImpl struct {
	logger             *zap.SugaredLogger
	HelmListClusterMap map[int]map[string]*client.DeployedAppDetail
	mutex              sync.Mutex
	informerStopper    map[int]chan struct{}
	clusterRepository  repository.ClusterRepository
	helmReleaseConfig  *globalConfig.HelmReleaseConfig
	k8sUtil            k8sUtils.K8sService
	listeners          []ClusterSecretUpdateListener
	converter          converter.ClusterBeanConverter
	runnable           *async.Runnable
}

func Newk8sInformerImpl(logger *zap.SugaredLogger, clusterRepository repository.ClusterRepository,
	helmReleaseConfig *globalConfig.HelmReleaseConfig, k8sUtil k8sUtils.K8sService,
	converter converter.ClusterBeanConverter, runnable *async.Runnable) (*K8sInformerImpl, error) {
	informerFactory := &K8sInformerImpl{
		logger:             logger,
		clusterRepository:  clusterRepository,
		helmReleaseConfig:  helmReleaseConfig,
		k8sUtil:            k8sUtil,
		converter:          converter,
		runnable:           runnable,
		HelmListClusterMap: make(map[int]map[string]*client.DeployedAppDetail),
		informerStopper:    make(map[int]chan struct{}),
	}
	if informerFactory.helmReleaseConfig.IsHelmReleaseCachingEnabled() {
		err := informerFactory.registerInformersForAllClusters()
		if err != nil {
			return nil, err
		}
	}
	return informerFactory, nil
}

func (impl *K8sInformerImpl) registerInformersForAllClusters() error {
	// for oss installation, there is race condition in starting informer,
	// where migration is not completed. So getting all active clusters prior
	// to ensure informer is started for all clusters
	models, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return err
	}
	clusterInfos := impl.converter.GetAllClusterInfo(models...)
	runnableFunc := func() {
		_ = impl.BuildInformerForAllClusters(clusterInfos)
	}
	impl.runnable.Execute(runnableFunc)
	return nil
}

func (impl *K8sInformerImpl) OnStateChange(clusterId int, action string) {
	impl.logger.Infow("syncing informer on cluster config update/delete", "action", action, "clusterId", clusterId)
	switch action {
	case UPDATE:
		err := impl.syncInformer(clusterId)
		if err != nil && !errors.Is(err, InformerAlreadyExistError) {
			impl.logger.Errorw("error in updating informer for cluster", "id", clusterId, "err", err)
			return
		}
	case DELETE:
		deleteClusterInfo, err := impl.clusterRepository.FindByIdWithActiveFalse(clusterId)
		if err != nil {
			impl.logger.Errorw("Error in fetching cluster by id", "cluster-id ", clusterId)
			return
		}
		impl.stopInformer(deleteClusterInfo.Id)
	}
}

func (impl *K8sInformerImpl) RegisterListener(listener ClusterSecretUpdateListener) {
	impl.logger.Infow("registering listener %s", reflect.TypeOf(listener))
	impl.listeners = append(impl.listeners, listener)
}

func (impl *K8sInformerImpl) BuildInformerForAllClusters(clusterInfos []*bean.ClusterInfo) error {
	if len(clusterInfos) == 0 {
		clusterInfo := &bean.ClusterInfo{
			ClusterId:             1,
			ClusterName:           DEFAULT_CLUSTER,
			InsecureSkipTLSVerify: true,
		}
		err := impl.startInformer(*clusterInfo)
		if err != nil {
			impl.logger.Errorw("error in starting informer for cluster ", "cluster-name ", clusterInfo.ClusterName, "err", err)
			return err
		}
		return nil
	}

	for _, clusterInfo := range clusterInfos {
		if clusterInfo == nil {
			// safety check
			continue
		}
		err := impl.startInformer(*clusterInfo)
		if err != nil {
			impl.logger.Errorw("error in starting informer for cluster ", "cluster-name ", clusterInfo.ClusterName, "err", err)
			// error state could be due to unreachable cluster, so continue with other clusters
		}
	}

	return nil
}

func (impl *K8sInformerImpl) GetClusterClientSet(clusterInfo bean.ClusterInfo) (*kubernetes.Clientset, error) {
	clusterConfig := impl.converter.GetClusterConfig(&clusterInfo)
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, err
	}
	httpClientFor, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while overriding k8s client", "reason", err)
		return nil, err
	}
	clusterClient, err := kubernetes.NewForConfigAndClient(restConfig, httpClientFor)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return nil, err
	}
	return clusterClient, nil
}

func (impl *K8sInformerImpl) startInformer(clusterInfo bean.ClusterInfo) error {
	clusterClient, err := impl.GetClusterClientSet(clusterInfo)
	if err != nil {
		impl.logger.Errorw("error in GetClusterClientSet", "clusterName", clusterInfo.ClusterName, "err", err)
		return err
	}

	// for default cluster adding an extra informer, this informer will add informer on new clusters
	if clusterInfo.ClusterName == DEFAULT_CLUSTER {
		impl.logger.Debugw("starting informer, reading new cluster request for default cluster")
		labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			//kubectl  get  secret --field-selector type==cluster.request/modify --all-namespaces
			opts.FieldSelector = "type==cluster.request/modify"
		})
		informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, 15*time.Minute, labelOptions)
		stopper := make(chan struct{})
		secretInformer := informerFactory.Core().V1().Secrets()
		_, err = secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				startTime := time.Now()
				impl.logger.Debugw("CLUSTER_ADD_INFORMER: cluster secret add event received", "obj", obj, "time", time.Now())
				if secretObject, ok := obj.(*coreV1.Secret); ok {
					if secretObject.Type != CLUSTER_MODIFY_EVENT_SECRET_TYPE {
						return
					}
					data := secretObject.Data
					action := data["action"]
					id := string(data["cluster_id"])
					var idInt int
					idInt, err = strconv.Atoi(id)
					if err != nil {
						impl.logger.Errorw("error in converting cluster id to int", "clusterId", id, "err", err)
						return
					}
					if string(action) == ADD {
						err = impl.startInformerAndPopulateCache(idInt)
						if err != nil && !errors.Is(err, InformerAlreadyExistError) {
							impl.logger.Errorw("error in adding informer for cluster", "id", idInt, "err", err)
							return
						}
					}
					if string(action) == UPDATE {
						err = impl.syncInformer(idInt)
						if err != nil && !errors.Is(err, InformerAlreadyExistError) {
							impl.logger.Errorw("error in updating informer for cluster", "id", clusterInfo.ClusterId, "name", clusterInfo.ClusterName, "err", err)
							return
						}
					}
					impl.logger.Infow("CLUSTER_ADD_INFORMER: registered informer for cluster", "clusterId", idInt, "timeTaken", time.Since(startTime))
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				startTime := time.Now()
				impl.logger.Debugw("CLUSTER_UPDATE_INFORMER: cluster secret update event received", "oldObj", oldObj, "newObj", newObj, "time", time.Now())
				if secretObject, ok := newObj.(*coreV1.Secret); ok {
					if secretObject.Type != CLUSTER_MODIFY_EVENT_SECRET_TYPE {
						return
					}
					data := secretObject.Data
					action := data["action"]
					id := string(data["cluster_id"])
					var idInt int
					idInt, err = strconv.Atoi(id)
					if err != nil {
						impl.logger.Errorw("error in converting cluster id to int", "clusterId", id, "err", err)
						return
					}
					if string(action) == ADD {
						err = impl.startInformerAndPopulateCache(idInt)
						if err != nil && !errors.Is(err, InformerAlreadyExistError) {
							impl.logger.Errorw("error in adding informer for cluster", "clusterId", idInt, "err", err)
							return
						}
					}
					if string(action) == UPDATE {
						impl.OnStateChange(idInt, string(action))
						impl.informOtherListeners(idInt, string(action))
					}
					impl.logger.Infow("CLUSTER_UPDATE_INFORMER: registered informer for cluster", "clusterId", idInt, "timeTaken", time.Since(startTime))
				}
			},
			DeleteFunc: func(obj interface{}) {
				startTime := time.Now()
				impl.logger.Debugw("CLUSTER_DELETE_INFORMER: secret delete event received", "obj", obj, "time", time.Now())
				if secretObject, ok := obj.(*coreV1.Secret); ok {
					if secretObject.Type != CLUSTER_MODIFY_EVENT_SECRET_TYPE {
						return
					}
					data := secretObject.Data
					action := data["action"]
					id := string(data["cluster_id"])
					idInt, err := strconv.Atoi(id)
					if err != nil {
						impl.logger.Errorw("error in converting cluster id to int", "clusterId", id, "err", err)
						return
					}
					if string(action) == DELETE {
						impl.OnStateChange(idInt, string(action))
						impl.informOtherListeners(idInt, string(action))
					}
					impl.logger.Infow("CLUSTER_DELETE_INFORMER: registered informer for cluster", "clusterId", idInt, "timeTaken", time.Since(startTime))
				}
			},
		})
		if err != nil {
			impl.logger.Errorw("error in adding event handler for cluster secret informer", "err", err)
			return err
		}
		informerFactory.Start(stopper)
		//impl.informerStopper[clusterInfo.ClusterName+"_second_informer"] = stopper

	}
	// these informers will be used to populate helm release cache

	err = impl.startInformerAndPopulateCache(clusterInfo.ClusterId)
	if err != nil && !errors.Is(err, InformerAlreadyExistError) {
		impl.logger.Errorw("error in creating informer for new cluster", "err", err)
		return err
	}

	return nil
}

func (impl *K8sInformerImpl) informOtherListeners(clusterId int, action string) {
	for _, listener := range impl.listeners {
		listener.OnStateChange(clusterId, action)
	}
}

func (impl *K8sInformerImpl) syncInformer(clusterId int) error {

	clusterInfo, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster info by id", "clusterId", clusterId, "err", err)
		return err
	}
	//before creating new informer for cluster, close existing one
	impl.logger.Debugw("stopping informer for cluster - ", "clusterName", clusterInfo.ClusterName, "clusterId", clusterInfo.Id)
	impl.stopInformer(clusterInfo.Id)
	impl.logger.Debugw("informer stopped", "clusterName", clusterInfo.ClusterName, "clusterId", clusterInfo.Id)
	//create new informer for cluster with new config
	err = impl.startInformerAndPopulateCache(clusterId)
	if err != nil {
		impl.logger.Errorw("error in starting informer for", "clusterName", clusterInfo.ClusterName)
		return err
	}
	return nil
}

func (impl *K8sInformerImpl) stopInformer(clusterId int) {
	if stopper, ok := impl.informerStopper[clusterId]; ok && stopper != nil {
		close(stopper)
		delete(impl.informerStopper, clusterId)
	}
	return
}

func (impl *K8sInformerImpl) transformHelmRelease(clusterModel *repository.Cluster, obj any) (*coreV1.Secret, error) {
	startTime := time.Now()
	if secretObject, ok := obj.(*coreV1.Secret); ok && secretObject.Type == HELM_RELEASE_SECRET_TYPE {
		releaseDTO, err := decodeHelmReleaseData(string(secretObject.Data["release"]))
		if err != nil {
			impl.logger.Error("error in decoding helm release", "clusterId", clusterModel.Id, "timeTaken", time.Since(startTime), "err", err)
			return nil, err
		}
		appDetail := adapter.ParseDeployedAppDetail(int32(clusterModel.Id), clusterModel.ClusterName, releaseDTO)
		transformedSecretData, err := parseSecretDataForDeployedAppDetail(appDetail)
		if err != nil {
			impl.logger.Error("error in parsing secret data for deployed app detail", "clusterId", clusterModel.Id, "timeTaken", time.Since(startTime), "err", err)
			return nil, err
		}
		secretObject.Data = transformedSecretData
		impl.logger.Debugw("successfully decoded helm release", "clusterId", clusterModel.Id, "timeTaken", time.Since(startTime))
		middleware.InformerDataTransformDuration.WithLabelValues(clusterModel.ClusterName, releaseDTO.Namespace, releaseDTO.Name).Observe(time.Since(startTime).Seconds())
		return secretObject, nil
	}
	impl.logger.Warnw("not a helm release secret", "clusterId", clusterModel.Id, "obj", obj)
	return nil, errors.New("error: not a helm release secret")
}

func (impl *K8sInformerImpl) startInformerAndPopulateCache(clusterId int) error {

	clusterModel, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster by cluster id", "clusterId", clusterId, "err", err)
		return err
	}

	if _, ok := impl.informerStopper[clusterId]; ok {
		impl.logger.Debugw(fmt.Sprintf("informer for %s already exist", clusterModel.ClusterName))
		return InformerAlreadyExistError
	}

	impl.logger.Infow("starting informer for cluster - ", "cluster-id ", clusterModel.Id, "cluster-name ", clusterModel.ClusterName)

	clusterInfo := impl.converter.GetClusterInfo(clusterModel)
	clusterConfig := impl.converter.GetClusterConfig(clusterInfo)
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "err", err, "clusterName", clusterConfig.ClusterName)
		return err
	}
	httpClientFor, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		impl.logger.Errorw("error occurred while overriding k8s client", "reason", err)
		return err
	}
	clusterClient, err := kubernetes.NewForConfigAndClient(restConfig, httpClientFor)
	if err != nil {
		impl.logger.Errorw("error in create k8s config", "err", err)
		return err
	}

	impl.mutex.Lock()
	impl.HelmListClusterMap[clusterId] = make(map[string]*client.DeployedAppDetail)
	impl.mutex.Unlock()

	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		//kubectl  get  secret --field-selector type==helm.sh/release.v1 -l status=deployed  --all-namespaces
		opts.LabelSelector = "status!=superseded"
		opts.FieldSelector = "type==helm.sh/release.v1"
	})
	transformerFunc := kubeinformers.WithTransform(func(obj any) (any, error) {
		return impl.transformHelmRelease(clusterModel, obj)
	})
	informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clusterClient, 15*time.Minute, labelOptions, transformerFunc)
	stopper := make(chan struct{})
	secretInformer := informerFactory.Core().V1().Secrets()
	_, err = secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			startTime := time.Now()
			impl.logger.Debugw("RELEASE_ADD_INFORMER: helm secret add event received", "clusterId", clusterModel.Id, "obj", obj, "time", time.Now())
			if secretObject, ok := obj.(*coreV1.Secret); ok {
				if secretObject == nil {
					impl.logger.Errorw("secret object is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				appDetail, err := getDeployedAppDetailFromSecretData(secretObject.Data)
				if err != nil {
					impl.logger.Errorw("error in getting deployed app detail from secret data", "clusterId", clusterModel.Id, "err", err)
					return
				}
				if appDetail == nil {
					impl.logger.Errorw("app detail is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				impl.mutex.Lock()
				defer impl.mutex.Unlock()
				impl.HelmListClusterMap[clusterId][impl.getUniqueReleaseKey(NewDeployedAppDetailDto(appDetail))] = appDetail
				impl.logger.Infow("RELEASE_ADD_INFORMER: added app detail in cache", "clusterId", clusterModel.Id, "namespace", appDetail.EnvironmentDetail.Namespace, "releaseName", appDetail.AppName, "timeTaken", time.Since(startTime))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			startTime := time.Now()
			impl.logger.Debugw("RELEASE_UPDATE_INFORMER: helm secret update event received", "clusterId", clusterModel.Id, "oldObj", oldObj, "newObj", newObj, "time", time.Now())
			if secretObject, ok := newObj.(*coreV1.Secret); ok {
				if secretObject == nil {
					impl.logger.Errorw("secret object is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				appDetail, err := getDeployedAppDetailFromSecretData(secretObject.Data)
				if err != nil {
					impl.logger.Errorw("error in getting deployed app detail from secret data", "clusterId", clusterModel.Id, "err", err)
					return
				}
				if appDetail == nil {
					impl.logger.Errorw("app detail is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				impl.mutex.Lock()
				defer impl.mutex.Unlock()
				impl.HelmListClusterMap[clusterId][impl.getUniqueReleaseKey(NewDeployedAppDetailDto(appDetail))] = appDetail
				impl.logger.Infow("RELEASE_UPDATE_INFORMER: updated app detail in cache", "clusterId", clusterModel.Id, "namespace", appDetail.EnvironmentDetail.Namespace, "releaseName", appDetail.AppName, "timeTaken", time.Since(startTime))
			}
		},
		DeleteFunc: func(obj interface{}) {
			startTime := time.Now()
			impl.logger.Debugw("RELEASE_DELETE_INFORMER: helm secret delete event received", "clusterId", clusterModel.Id, "obj", obj, "time", time.Now())
			if secretObject, ok := obj.(*coreV1.Secret); ok {
				if secretObject == nil {
					impl.logger.Errorw("secret object is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				appDetail, err := getDeployedAppDetailFromSecretData(secretObject.Data)
				if err != nil {
					impl.logger.Errorw("error in getting deployed app detail from secret data", "clusterId", clusterModel.Id, "err", err)
					return
				}
				if appDetail == nil {
					impl.logger.Errorw("app detail is nil! unexpected...", "clusterId", clusterModel.Id)
					return
				}
				impl.mutex.Lock()
				defer impl.mutex.Unlock()
				delete(impl.HelmListClusterMap[clusterId], impl.getUniqueReleaseKey(NewDeployedAppDetailDto(appDetail)))
				impl.logger.Infow("RELEASE_DELETE_INFORMER: deleted app detail in cache", "clusterId", clusterModel.Id, "namespace", appDetail.EnvironmentDetail.Namespace, "releaseName", appDetail.AppName, "timeTaken", time.Since(startTime))
			}
		},
	})
	if err != nil {
		impl.logger.Errorw("error in adding event handler for helm secret informer", "clusterId", clusterId, "err", err)
		return err
	}
	informerFactory.Start(stopper)
	impl.logger.Infow("informer started for cluster: ", "clusterId", clusterModel.Id, "clusterName", clusterModel.ClusterName)
	impl.informerStopper[clusterId] = stopper
	return nil
}

func (impl *K8sInformerImpl) getUniqueReleaseKey(appDetailDto *DeployedAppDetailDto) string {
	return appDetailDto.getUniqueReleaseIdentifier()
}

func (impl *K8sInformerImpl) GetAllReleaseByClusterId(clusterId int) []*client.DeployedAppDetail {

	var deployedAppDetailList []*client.DeployedAppDetail
	releaseMap := impl.HelmListClusterMap[clusterId]
	for _, v := range releaseMap {
		deployedAppDetailList = append(deployedAppDetailList, v)
	}
	return deployedAppDetailList
}

func (impl *K8sInformerImpl) CheckReleaseExists(clusterId int32, releaseIdentifier string) bool {
	releaseMap := impl.HelmListClusterMap[int(clusterId)]
	_, ok := releaseMap[releaseIdentifier]
	if ok {
		return true
	}
	return false
}

func (impl *K8sInformerImpl) GetReleaseDetails(clusterId int32, releaseIdentifier string) (*client.DeployedAppDetail, error) {
	releaseMap := impl.HelmListClusterMap[int(clusterId)]
	deployDetail, ok := releaseMap[releaseIdentifier]
	if !ok {
		return nil, ErrorCacheMissReleaseNotFound
	}
	return deployDetail, nil
}
