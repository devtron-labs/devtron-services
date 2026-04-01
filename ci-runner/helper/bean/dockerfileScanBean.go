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

package bean

// DockerfileScanRequest represents the request to scan a Dockerfile
type DockerfileScanRequest struct {
	AppId                 int    `json:"appId"`
	BuildId               int    `json:"buildId"`
	PipelineId            int    `json:"pipelineId"`
	DockerfileContent     string `json:"dockerfileContent"`
	DockerfileScanEnabled bool   `json:"dockerfileScanEnabled"`
	ForceDockerfileScan   bool   `json:"forceDockerfileScan"`
}

// ScanConfig holds configuration for Dockerfile scanning
type ScanConfig struct {
	ImageScannerEndpoint string `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-service.devtroncd:80"`
	FailOnError          bool   `env:"DOCKERFILE_SCAN_FAIL_ON_ERROR" envDefault:"false"`
	MaxRetries           int    `env:"DOCKERFILE_SCAN_MAX_RETRIES" envDefault:"3"`
	RetryWaitTimeSeconds int    `env:"DOCKERFILE_SCAN_RETRY_WAIT_SECONDS" envDefault:"5"`
}
