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
	"github.com/devtron-labs/ci-runner/util"
	"github.com/go-resty/resty/v2"
)

// DockerfileScanRequest represents the request to scan a Dockerfile
type DockerfileScanRequest struct {
	AppId                 int      `json:"appId"`
	BuildId               int      `json:"buildId"`
	PipelineId            int      `json:"pipelineId"`
	DockerfileContent     string   `json:"dockerfileContent"`
	DockerfileScanEnabled bool     `json:"dockerfileScanEnabled"`
	ForceDockerfileScan   bool     `json:"forceDockerfileScan"`
	IgnoredRules          []string `json:"ignoredRules"`
}

// ScanConfig holds configuration for Dockerfile scanning
type ScanConfig struct {
	ImageScannerEndpoint string `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-service.devtroncd:80"`
	FailOnError          bool   `env:"DOCKERFILE_SCAN_FAIL_ON_ERROR" envDefault:"false"`
	MaxRetries           int    `env:"DOCKERFILE_SCAN_MAX_RETRIES" envDefault:"3"`
	RetryWaitTimeSeconds int    `env:"DOCKERFILE_SCAN_RETRY_WAIT_SECONDS" envDefault:"5"`
}

// MaxDockerfileSize is the maximum allowed Dockerfile size (1MB)
const MaxDockerfileSize = 1 * 1024 * 1024 // 1MB

// InitiateDockerfileScan initiates a Dockerfile scan using hadolint
// It reads the Dockerfile from filesystem and sends content to image-scanner service
// Note: The decision to run the scan is made by the caller (runBuildArtifact)
// which handles FORCE_DOCKERFILE_SCAN flag and pipeline-level DockerfileScanEnabled settings
func InitiateDockerfileScan(ciRequest *CommonWorkflowRequest) {
	log.Println(util.DEVTRON, "initiating Dockerfile scan")

	// Validate config exists
	if ciRequest.CiBuildConfig == nil || ciRequest.CiBuildConfig.DockerBuildConfig == nil {
		log.Println(util.DEVTRON, "docker build config not found, skipping Dockerfile scan")
		return
	}

	// Wait for git clone to complete (checkout path to exist)
	// Use the SAME path resolution as Docker build (getDockerfilePath)
	var dockerfilePath string
	if ciRequest.CiBuildConfig.CiBuildType == "managed-dockerfile-build" {
		// For managed Dockerfile, use GetSelfManagedDockerfilePath
		dockerfilePath = filepath.Join(util.WORKINGDIR, ciRequest.CheckoutPath, "./Dockerfile")
	} else {
		// For self-managed Dockerfile, use the configured path
		dockerfilePath = ciRequest.CiBuildConfig.DockerBuildConfig.DockerfilePath
	}
	// Convert to absolute path (same as Docker build)
	dockerfilePath, _ = filepath.Abs(dockerfilePath)

	// Fallback wait (should not be needed - scan triggered at Docker build start)
	maxWait := 2 * time.Minute
	waitInterval := 10 * time.Second
	startTime := time.Now()

	log.Println(util.DEVTRON, "dockerfile scan: waiting for git clone to complete", "path", dockerfilePath, "buildId", ciRequest.WorkflowId)

	for time.Since(startTime) < maxWait {
		if _, err := os.Stat(dockerfilePath); err == nil {
			log.Println(util.DEVTRON, "dockerfile scan: Dockerfile found, proceeding", "path", dockerfilePath, "elapsed", time.Since(startTime).Round(time.Second))
			break // File exists, proceed
		}
		// Log progress every 30 seconds
		if int(time.Since(startTime).Seconds())%30 == 0 {
			log.Println(util.DEVTRON, "dockerfile scan: waiting for git clone to complete...", "path", dockerfilePath, "elapsed", time.Since(startTime).Round(time.Second), "maxWait", maxWait)
		}
		time.Sleep(waitInterval)
	}

	// Read Dockerfile from filesystem (single source of truth)
	dockerfileContent, err := os.ReadFile(dockerfilePath)
	if err != nil {
		log.Println(util.DEVTRON, "error in reading Dockerfile for scanning",
			"path", dockerfilePath, "err", err)
		if err := handleScanError(fmt.Sprintf("Failed to read Dockerfile from %s: %v", dockerfilePath, err), ciRequest.DockerfileScanEnabled); err != nil {
			return
		}
		return
	}

	// Prepare scan request with Dockerfile content
	// CRITICAL FIX: Read DockerfileScanEnabled from ciRequest.DockerfileScanEnabled (CommonWorkflowRequest level)
	// NOT from ciRequest.CiBuildConfig.DockerBuildConfig.DockerfileScanEnabled (which may be out of sync)
	// ForceDockerfileScan is only available at CommonWorkflowRequest level
	scanRequest := &DockerfileScanRequest{
		AppId:                 ciRequest.AppId,
		BuildId:               ciRequest.WorkflowId,
		PipelineId:            ciRequest.PipelineId,
		DockerfileContent:     string(dockerfileContent),
		DockerfileScanEnabled: ciRequest.DockerfileScanEnabled,
		ForceDockerfileScan:   ciRequest.ForceDockerfileScan,
		IgnoredRules:          []string{}, // Can be populated from config in future
	}

	jsonBody, err := json.Marshal(scanRequest)
	if err != nil {
		log.Println(util.DEVTRON, "error in marshalling Dockerfile scan request", "err", err)
		if err := handleScanError(fmt.Sprintf("Failed to marshal scan request: %v", err), ciRequest.DockerfileScanEnabled); err != nil {
			return
		}
		return
	}

	cfg := &ScanConfig{}
	err = env.Parse(cfg)
	if err != nil {
		log.Println(util.DEVTRON, "error in parsing scan config", "err", err)
		if err := handleScanError(fmt.Sprintf("Failed to parse scan config: %v", err), ciRequest.DockerfileScanEnabled); err != nil {
			return
		}
		return
	}

	// Create HTTP client with timeout and configurable retries
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

	// Record success/failure in circuit breaker
	if err != nil || (resp != nil && (resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK)) {
		log.Println(util.DEVTRON, "circuit breaker recorded FAILURE", "buildId", ciRequest.WorkflowId)
	} else {
		log.Println(util.DEVTRON, "circuit breaker recorded SUCCESS", "buildId", ciRequest.WorkflowId)
	}

	if err != nil {
		log.Println(util.DEVTRON, "error in calling image-scanner for Dockerfile scan", "err", err)
		if err := handleScanError(fmt.Sprintf("Dockerfile scan failed: %v", err), cfg.FailOnError); err != nil {
			return
		}
		return
	}

	// Accept both 202 (Accepted) and 200 (OK - for cached results)
	if resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusOK {
		log.Println(util.DEVTRON, "image-scanner returned non-202/200 status for Dockerfile scan",
			"status", resp.StatusCode(), "body", string(resp.Body()))
		if err := handleScanError(fmt.Sprintf("Dockerfile scan failed with status: %d", resp.StatusCode()), cfg.FailOnError); err != nil {
			return
		}
		return
	}

	log.Println(util.DEVTRON, "successfully initiated Dockerfile scan",
		"statusCode", resp.StatusCode(), "buildId", ciRequest.WorkflowId)
}

// handleScanError handles scan errors based on FailOnError configuration
func handleScanError(message string, failOnError bool) error {
	if failOnError {
		log.Println(util.DEVTRON, "Dockerfile scan failed (fail-on-error enabled)", "message", message)
		return fmt.Errorf("Dockerfile scan failed: %s", message)
	}
	// Log warning but don't fail the build
	log.Println(util.DEVTRON, "Dockerfile scan failed (fail-on-error disabled)", "message", message)
	return nil
}
