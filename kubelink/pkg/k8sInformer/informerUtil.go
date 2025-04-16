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
	"encoding/json"
	"errors"
	client "github.com/devtron-labs/kubelink/grpc"
)

func parseSecretDataForDeployedAppDetail(appDetail *client.DeployedAppDetail) (map[string][]byte, error) {
	appDetailBytes, err := json.Marshal(appDetail)
	if err != nil {
		return nil, err
	}
	data := make(map[string][]byte)
	data[secretKeyAppDetailKey] = appDetailBytes
	return data, nil
}

func getDeployedAppDetailFromSecretData(data map[string][]byte) (*client.DeployedAppDetail, error) {
	if appDetailBytes, ok := data[secretKeyAppDetailKey]; ok {
		var appDetail client.DeployedAppDetail
		err := json.Unmarshal(appDetailBytes, &appDetail)
		if err != nil {
			return nil, err
		}
		return &appDetail, nil
	}
	return nil, errors.New("app detail not found in secret data")
}
