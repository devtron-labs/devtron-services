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
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/service/helmApplicationService/release"
	"io"
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

func decodeHelmReleaseData(data string) (*release.Release, error) {
	// base64 decode string
	b64 := base64.StdEncoding
	b, err := b64.DecodeString(data)
	if err != nil {
		return nil, err
	}

	var magicGzip = []byte{0x1f, 0x8b, 0x08}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if len(b) > 3 && bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		b2, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls release.Release
	// unmarshal release object bytes
	if err := json.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}
