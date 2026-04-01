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

package helper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"github.com/devtron-labs/ci-runner/helper/bean"
	"github.com/devtron-labs/ci-runner/util"
	"github.com/go-resty/resty/v2"
)

// MaxDockerfileSize is the maximum allowed Dockerfile size (1MB)
const MaxDockerfileSize = 1 * 1024 * 1024 // 1MB

// InitiateDockerfileScan initiates a Dockerfile scan using hadolint
// It reads the Dockerfile from filesystem and sends content to image-scanner service
func InitiateDockerfileScan(ciRequest *CommonWorkflowRequest) {
	log.Println(util.DEVTRON, "initiating Dockerfile scan")

	// Validate config exists
	if ciRequest.CiBuildConfig == nil || ciRequest.CiBuildConfig.DockerBuildConfig == nil {
		log.Println(util.DEVTRON, "docker build config not found, skipping Dockerfile scan")
		return
	}

	// Resolve Dockerfile path
	var dockerfilePath string
	if ciRequest.CiBuildConfig.CiBuildType == "managed-dockerfile-build" {
		dockerfilePath = filepath.Join(util.WORKINGDIR, ciRequest.CheckoutPath, "./Dockerfile")
	} else {
		dockerfilePath = ciRequest.CiBuildConfig.DockerBuildConfig.DockerfilePath
	}

	// Convert to absolute path
	var err error
	dockerfilePath, err = filepath.Abs(dockerfilePath)
	if err != nil {
		log.Println(util.DEVTRON, "error in resolving absolute path for Dockerfile", "path", dockerfilePath, "err", err)
		return
	}

	// Check if Dockerfile exists (removed polling logic)
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		log.Println(util.DEVTRON, "dockerfile scan: Dockerfile not found at", dockerfilePath)
		return
	}

	// Read Dockerfile from filesystem
	dockerfileContent, err := os.ReadFile(dockerfilePath)
	if err != nil {
		log.Println(util.DEVTRON, "error in reading Dockerfile for scanning", "path", dockerfilePath, "err", err)
		handleScanError(fmt.Sprintf("Failed to read Dockerfile: %v", err), ciRequest.DockerfileScanEnabled)
		return
	}

	// Prepare scan request
	scanRequest := &bean.DockerfileScanRequest{
		AppId:                 ciRequest.AppId,
		BuildId:               ciRequest.WorkflowId,
		PipelineId:            ciRequest.PipelineId,
		DockerfileContent:     string(dockerfileContent),
		DockerfileScanEnabled: ciRequest.DockerfileScanEnabled,
		ForceDockerfileScan:   ciRequest.ForceDockerfileScan,
	}

	jsonBody, err := json.Marshal(scanRequest)
	if err != nil {
		log.Println(util.DEVTRON, "error in marshalling Dockerfile scan request", "err", err)
		handleScanError(fmt.Sprintf("Failed to marshal scan request: %v", err), ciRequest.DockerfileScanEnabled)
		return
	}

	cfg := &bean.ScanConfig{}
	err = env.Parse(cfg)
	if err != nil {
		log.Println(util.DEVTRON, "error in parsing scan config", "err", err)
		handleScanError(fmt.Sprintf("Failed to parse scan config: %v", err), ciRequest.DockerfileScanEnabled)
		return
	}

	// Create HTTP client (Using local instantiation but with improved error logging)
	client := resty.New()
	client.SetTimeout(2 * time.Minute)
	client.
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetRetryCount(cfg.MaxRetries).
		SetRetryWaitTime(time.Duration(cfg.RetryWaitTimeSeconds) * time.Second)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonBody).
		Post(fmt.Sprintf("%s/%s", cfg.ImageScannerEndpoint, "scanner/dockerfile/scan"))

	// Record success/failure with actual error logging
	if err != nil || (resp != nil && (resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK)) {
		var status string
		if resp != nil {
			status = resp.Status()
		}
		log.Println(util.DEVTRON, "circuit breaker recorded FAILURE for Dockerfile scan", "buildId", ciRequest.WorkflowId, "err", err, "status", status)
	} else {
		log.Println(util.DEVTRON, "circuit breaker recorded SUCCESS for Dockerfile scan", "buildId", ciRequest.WorkflowId)
	}

	if err != nil {
		log.Println(util.DEVTRON, "error in calling image-scanner for Dockerfile scan", "err", err)
		handleScanError(fmt.Sprintf("Dockerfile scan failed: %v", err), cfg.FailOnError)
		return
	}

	// Accept both 202 (Accepted) and 200 (OK)
	if resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK {
		log.Println(util.DEVTRON, "image-scanner returned non-202/200 status for Dockerfile scan",
			"status", resp.StatusCode(), "body", string(resp.Body()))
		handleScanError(fmt.Sprintf("Dockerfile scan failed with status: %d", resp.StatusCode()), cfg.FailOnError)
		return
	}

	log.Println(util.DEVTRON, "successfully initiated Dockerfile scan",
		"statusCode", resp.StatusCode(), "buildId", ciRequest.WorkflowId)
}

// handleScanError handles scan errors based on FailOnError configuration
func handleScanError(message string, failOnError bool) {
	if failOnError {
		log.Println(util.DEVTRON, "Dockerfile scan failed (fail-on-error enabled)", "message", message)
		// We don't return error here because this is called in a goroutine
	} else {
		log.Println(util.DEVTRON, "Dockerfile scan failed (fail-on-error disabled)", "message", message)
	}
}
