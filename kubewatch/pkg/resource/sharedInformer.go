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

package resource

import (
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
	"github.com/devtron-labs/kubewatch/pkg/resource/application"
	"github.com/devtron-labs/kubewatch/pkg/resource/bean"
	veleroBackup "github.com/devtron-labs/kubewatch/pkg/resource/veleroResource/backup"
	veleroBackupSchedule "github.com/devtron-labs/kubewatch/pkg/resource/veleroResource/backupSchedule"
	veleroBSL "github.com/devtron-labs/kubewatch/pkg/resource/veleroResource/bsl"
	veleroRestore "github.com/devtron-labs/kubewatch/pkg/resource/veleroResource/restore"
	veleroVSL "github.com/devtron-labs/kubewatch/pkg/resource/veleroResource/vsl"
	"github.com/devtron-labs/kubewatch/pkg/resource/workflow"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type SharedInformer interface {
	GetSharedInformer(clusterLabels *informerBean.ClusterLabels, namespace string, k8sConfig *rest.Config) (cache.SharedIndexInformer, error)
}

func (impl *InformerClientImpl) GetSharedInformerClient(sharedInformerType bean.SharedInformerType) SharedInformer {
	switch sharedInformerType {
	case bean.ApplicationResourceType:
		return application.NewInformerImpl(impl.logger, impl.client)
	case bean.CiWorkflowResourceType:
		return workflow.NewCiInformerImpl(impl.logger, impl.client, impl.appConfig)
	case bean.CdWorkflowResourceType:
		return workflow.NewCdInformerImpl(impl.logger, impl.client, impl.appConfig)
	case bean.VeleroBslResourceType:
		return veleroBSL.NewInformerImpl(impl.logger, impl.client)
	case bean.VeleroVslResourceType:
		return veleroVSL.NewInformerImpl(impl.logger, impl.client)
	case bean.VeleroBackupResourceType:
		return veleroBackup.NewInformerImpl(impl.logger, impl.client)
	case bean.VeleroRestoreResourceType:
		return veleroRestore.NewInformerImpl(impl.logger, impl.client)
	case bean.VeleroBackupScheduleResourceType:
		return veleroBackupSchedule.NewInformerImpl(impl.logger, impl.client)

	default:
		return NewUnimplementedImpl()
	}
}
