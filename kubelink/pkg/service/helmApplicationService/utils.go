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
package helmApplicationService

import (
	"bytes"
	"errors"
	"fmt"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/helmClient"
	"github.com/devtron-labs/kubelink/pkg/k8sInformer"
	chart2 "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	downloader2 "helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

func getUniqueReleaseIdentifierName(releaseIdentifier *client.ReleaseIdentifier) string {
	return releaseIdentifier.ReleaseNamespace + "_" + releaseIdentifier.ReleaseName + "_" + strconv.Itoa(int(releaseIdentifier.ClusterConfig.ClusterId))
}

const (
	DirCreatingError = "err in creating dir"
)

func IsReleaseNotFoundInCacheError(err error) bool {
	return errors.Is(err, k8sInformer.ErrorCacheMissReleaseNotFound)
}

// UpdateChartLock updates the Chart.lock file based on its contents.
func UpdateChartLock(lockFilePath string, helmRelease *release.Release) error {
	lockFileData, err := os.ReadFile(lockFilePath)
	if err != nil {
		return fmt.Errorf("error reading Chart.lock file at %s: %w", lockFilePath, err)
	}
	helmRelease.Chart.Lock = &chart2.Lock{}
	if err := yaml.Unmarshal(lockFileData, helmRelease.Chart.Lock); err != nil {
		return fmt.Errorf("error unmarshalling Chart.lock file at %s: %w", lockFilePath, err)
	}
	return nil
}

// ProcessTGZFiles locates and processes .tgz files in the charts directory.
func ProcessTGZFiles(chartsDir string, helmRelease *release.Release) error {
	files, err := os.ReadDir(chartsDir)
	if err != nil {
		return fmt.Errorf("error reading charts directory in dir %s :%w", chartsDir, err)

	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tgz") {
			tgzPath := filepath.Join(chartsDir, file.Name())

			if err := expandAndSetDependency(tgzPath, helmRelease); err != nil {
				return err
			}
		}
	}
	return nil
}

// expandAndSetDependency expands a .tgz file and sets it as a dependency.
func expandAndSetDependency(tgzPath string, helmRelease *release.Release) error {
	// Create a temporary directory for expanding the chart
	tempDir, err := os.MkdirTemp("", "helmDependencyChart*")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer func() {
		// Clean up the temporary directory
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Sprintf("error in removing temporary directory %s: %s", tempDir, removeErr.Error())
		}
	}()

	// Open the .tgz file
	tgzFile, err := os.Open(tgzPath)
	if err != nil {
		return fmt.Errorf("error opening tgz file of tgzPath %s: %w", tgzPath, err)
	}
	defer func(tgzFile *os.File) {
		err := tgzFile.Close()
		if err != nil {
			fmt.Errorf("error in closing tgz file of tgzPath %s: %w", tgzPath, err)
		}
	}(tgzFile)

	// Expand the .tgz file into the temporary directory
	if err := chartutil.Expand(tempDir, tgzFile); err != nil {
		return fmt.Errorf("error expanding tgz file having filePath %s : %w", tgzPath, err)
	}

	// Dynamically find the expanded chart directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("error reading contents of temporary directory %s: %w", tempDir, err)
	}

	var expandedChartDir string
	for _, entry := range entries {
		if entry.IsDir() {
			expandedChartDir = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if expandedChartDir == "" {
		err := fmt.Errorf("no expanded chart directory found in %s", tempDir)
		return fmt.Errorf("error locating expanded chart directory found in %s: %w", tempDir, err)
	}

	// Load the expanded chart
	expandedChart, err := loader.LoadDir(expandedChartDir)
	if err != nil {
		return fmt.Errorf("error loading expanded chart : %w", err)
	}
	helmRelease.Chart.SetDependencies(expandedChart)
	return nil
}

func UpdateChartDependencies(providers getter.Providers, helmRelease *release.Release, registry *registry.Client) error {
	// Step 1: Update chart dependencies
	outputChartPathDir := fmt.Sprintf("%s", helmClient.DefaultTempDirectory)
	err := os.MkdirAll(outputChartPathDir, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(outputChartPathDir)
		if err != nil {
			fmt.Println("error in deleting dir", " dir: ", outputChartPathDir, " err: ", err)
		}
	}()
	abpath, err := chartutil.Save(helmRelease.Chart, outputChartPathDir)
	if err != nil {
		fmt.Println("error in saving chartdata in the destination dir ", " dir : ", outputChartPathDir, " err : ", err)
		return err
	}

	// Unpack the .tgz file to a directory
	h, err := os.Open(abpath)
	if err != nil {
		return err
	}
	if err := chartutil.Expand(helmClient.DefaultTempDirectory, h); err != nil {
		fmt.Println("error unpacking chart", "dir:", abpath, "err:", err)
		return err
	}
	outputChartPathDir = filepath.Join(helmClient.DefaultTempDirectory, helmRelease.Chart.Metadata.Name)
	outputBuffer := bytes.NewBuffer(nil)
	manager := &downloader2.Manager{
		ChartPath:        outputChartPathDir,
		Out:              outputBuffer,
		Getters:          providers,
		RepositoryConfig: helmClient.DefaultRepositoryConfigPath,
		RepositoryCache:  helmClient.DefaultCachePath,
		RegistryClient:   registry,
	}
	err = manager.Update()
	if err != nil {
		return fmt.Errorf("error updating chart dependencies: %w", err)
	}

	// Step 2: Check and process .tgz files in charts directory
	chartsDir := filepath.Join(outputChartPathDir, "charts")
	if err := ProcessTGZFiles(chartsDir, helmRelease); err != nil {
		return err
	}

	// Step 3: Update the Chart.lock file if it exists
	lockFilePath := filepath.Join(outputChartPathDir, "Chart.lock")
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		fmt.Sprintf("No Chart.lock file found, skipping lock file update, %s", lockFilePath)
		return nil
	}

	if err := UpdateChartLock(lockFilePath, helmRelease); err != nil {
		return err
	}
	return nil
}
