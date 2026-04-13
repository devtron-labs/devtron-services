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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/devtron-labs/ci-runner/helper/bean"
	"github.com/devtron-labs/ci-runner/util"
)

func TestInitiateDockerfileScan_Success(t *testing.T) {
	// Create a temporary directory structure that mimics the runner's environment
	// We use os.MkdirTemp but then we need to ensure the path aligns with util.WORKINGDIR
	// Since util.WORKINGDIR is constant "/devtroncd", we will create a temp dir and symlink or just use a unique path
	// To avoid permission issues, we will check if we can write to util.WORKINGDIR

	testDir := filepath.Join(util.WORKINGDIR, "test-scan-"+t.Name())
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Skipf("Skipping test: cannot create directory %s: %v", testDir, err)
	}
	defer os.RemoveAll(testDir)

	// Create mock Dockerfile
	dockerfileContent := "FROM node:18-alpine"
	dockerfilePath := filepath.Join(testDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock Server
	receivedRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequest = true
		var req bean.DockerfileScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.DockerfileContent != dockerfileContent {
			t.Errorf("Expected dockerfile content %s, got %s", dockerfileContent, req.DockerfileContent)
		}
		if !req.DockerfileScanEnabled {
			t.Error("Expected DockerfileScanEnabled to be true")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	// Setup Env
	os.Setenv("IMAGE_SCANNER_ENDPOINT", server.URL)
	os.Setenv("DOCKERFILE_SCAN_FAIL_ON_ERROR", "false")

	// Setup Request
	// CheckoutPath should be relative to WORKINGDIR, so "test-scan-..."
	req := &CommonWorkflowRequest{
		WorkflowId:            999,
		AppId:                 1,
		PipelineId:            2,
		DockerfileScanEnabled: true,
		CheckoutPath:          filepath.Base(testDir),
		CiBuildConfig: &CiBuildConfigBean{
			CiBuildType: "managed-dockerfile-build",
			DockerBuildConfig: &DockerBuildConfig{
				DockerfilePath: "./Dockerfile",
			},
		},
	}

	// Call
	InitiateDockerfileScan(req)

	if !receivedRequest {
		t.Error("Expected scan request to be sent to server")
	}
}

func TestInitiateDockerfileScan_MissingFile(t *testing.T) {
	testDir := filepath.Join(util.WORKINGDIR, "test-scan-missing-"+t.Name())
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Skipf("Skipping test: cannot create directory %s: %v", testDir, err)
	}
	defer os.RemoveAll(testDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called if Dockerfile is missing")
	}))
	defer server.Close()

	os.Setenv("IMAGE_SCANNER_ENDPOINT", server.URL)

	req := &CommonWorkflowRequest{
		WorkflowId:            999,
		AppId:                 1,
		PipelineId:            2,
		DockerfileScanEnabled: true,
		CheckoutPath:          filepath.Base(testDir),
		CiBuildConfig: &CiBuildConfigBean{
			CiBuildType: "managed-dockerfile-build",
		},
	}

	// Should return silently
	InitiateDockerfileScan(req)
}

func TestInitiateDockerfileScan_ServerError(t *testing.T) {
	testDir := filepath.Join(util.WORKINGDIR, "test-scan-error-"+t.Name())
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Skipf("Skipping test: cannot create directory %s: %v", testDir, err)
	}
	defer os.RemoveAll(testDir)

	if err := os.WriteFile(filepath.Join(testDir, "Dockerfile"), []byte("FROM node"), 0644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("IMAGE_SCANNER_ENDPOINT", server.URL)
	os.Setenv("DOCKERFILE_SCAN_FAIL_ON_ERROR", "false")

	req := &CommonWorkflowRequest{
		WorkflowId:            999,
		AppId:                 1,
		PipelineId:            2,
		DockerfileScanEnabled: true,
		CheckoutPath:          filepath.Base(testDir),
		CiBuildConfig: &CiBuildConfigBean{
			CiBuildType: "managed-dockerfile-build",
		},
	}

	// Should handle error gracefully
	InitiateDockerfileScan(req)
}
