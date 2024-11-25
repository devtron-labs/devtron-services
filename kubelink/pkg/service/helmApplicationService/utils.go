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
	"errors"
	"fmt"
	client "github.com/devtron-labs/kubelink/grpc"
	"github.com/devtron-labs/kubelink/pkg/k8sInformer"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"os"
	"path/filepath"
	"strconv"
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

// expandAndSetDependency expands a .tgz file and sets it as a dependency.
func expandAndSetDependency(tgzPath string, helmRelease *release.Release) error {
	// Create a temporary directory for expanding the chart
	tempDir, err := os.MkdirTemp("", "helmDependencyChart*")
	if err != nil {
		fmt.Errorf("error creating temporary directory: %w", err)
		return err
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
