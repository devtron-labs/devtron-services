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
	"github.com/devtron-labs/kubelink/pkg/k8sInformer/bean"
	jsoniter "github.com/json-iterator/go"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"time"
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

func decodeRelease(data string) (*release.Release, error) {
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

func decodeReleaseWithJsonIterator(data string) (*release.Release, error) {
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
	if err := jsoniter.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func decodeReleaseIntoCustomBean(data string) (*bean.Release, error) {
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

	var rls bean.Release
	// unmarshal release object bytes
	if err := json.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func decodeReleaseIntoCustomBeanWithJsonIterator(data string) (*bean.Release, error) {
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

	var rls bean.Release
	// unmarshal release object bytes
	if err := jsoniter.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func decodeReleaseIntoMap(data string) (*bean.Release, error) {
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

	var rlsMap map[string]interface{}
	// unmarshal release object bytes
	if err = json.Unmarshal(b, &rlsMap); err != nil {
		return nil, err
	}
	var rlsChart *bean.Chart
	if chartRaw, ok := rlsMap["chart"]; ok {
		if chartMap, ok := chartRaw.(map[string]interface{}); ok {
			rlsChart = &bean.Chart{}
			if metadataRaw, ok := chartMap["metadata"]; ok {
				if metadataMap, ok := metadataRaw.(map[string]interface{}); ok {
					rlsChart.Metadata = &chart.Metadata{}
					if name, ok := metadataMap["name"]; ok {
						rlsChart.Metadata.Name = name.(string)
					}
					if version, ok := metadataMap["version"]; ok {
						rlsChart.Metadata.Version = version.(string)
					}
					if icon, ok := metadataMap["icon"]; ok {
						rlsChart.Metadata.Icon = icon.(string)
					}
					if home, ok := metadataMap["home"]; ok {
						rlsChart.Metadata.Home = home.(string)
					}
				}
			}
		}
	}
	var rlsInfo *release.Info
	if infoRaw, ok := rlsMap["info"]; ok {
		if infoMap, ok := infoRaw.(map[string]interface{}); ok {
			rlsInfo = &release.Info{}
			if status, ok := infoMap["status"]; ok {
				rlsInfo.Status = release.Status(status.(string))
			}
			if lastDeployed, ok := infoMap["last_deployed"]; ok {
				lastDeployedTime, ok := lastDeployed.(string)
				if !ok {
					return nil, errors.New("last deployed time not found in release info")
				}
				lastDeployedTimeObj, err := time.Parse(time.RFC3339, lastDeployedTime)
				if err != nil {
					return nil, err
				}
				rlsInfo.LastDeployed.Time = lastDeployedTimeObj
			}
		}
	}
	rls := bean.Release{
		Chart:     rlsChart,
		Info:      rlsInfo,
		Name:      rlsMap["name"].(string),
		Namespace: rlsMap["namespace"].(string),
	}
	return &rls, nil
}
