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
	client "github.com/devtron-labs/kubelink/grpc"
)

type DeployedAppDetailDto struct {
	*client.DeployedAppDetail
}

func NewDeployedAppDetailDto(appDetail *client.DeployedAppDetail) *DeployedAppDetailDto {
	return &DeployedAppDetailDto{DeployedAppDetail: appDetail}
}

func (r *DeployedAppDetailDto) getUniqueReleaseIdentifier() string {
	if r == nil || r.EnvironmentDetail == nil {
		return ""
	}
	// adding cluster id with release name and namespace because there can be case when two cluster or two namespaces have release with same name
	return fmt.Sprintf("%s_%s_%d", r.EnvironmentDetail.Namespace, r.AppName, r.EnvironmentDetail.ClusterId)
}

var (
	ErrorCacheMissReleaseNotFound = errors.New("release not found in cache")
	InformerAlreadyExistError     = errors.New(INFORMER_ALREADY_EXIST_MESSAGE)
)
