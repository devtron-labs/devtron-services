/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"bytes"
	"github.com/devtron-labs/chart-sync/internals/sql"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"net/url"
	"time"
)

func GetFromUrlWithRetry(baseurl string, absoluteUrl string, username string, password string, allowInsecureConnection bool) (*bytes.Buffer, error) {
	var (
		err, errInGetUrl error
		response         *bytes.Buffer
		retries          = 3
	)
	getters := getter.All(&cli.EnvSettings{})
	u, err := url.Parse(absoluteUrl)
	if err != nil {
		return nil, errors.Errorf("invalid chart URL format: %s", baseurl)
	}

	client, err := getters.ByScheme(u.Scheme)

	if err != nil {
		return nil, errors.Errorf("could not find protocol handler for: %s", u.Scheme)
	}

	for retries > 0 {

		var options []getter.Option

		if allowInsecureConnection {
			options = append(options, getter.WithInsecureSkipVerifyTLS(allowInsecureConnection))
		}
		if len(username) > 0 && len(password) > 0 {
			options = append(options, getter.WithBasicAuth(username, password))
		}
		if len(baseurl) > 0 {
			options = append(options, getter.WithURL(baseurl))
		}

		response, errInGetUrl = client.Get(absoluteUrl, options...)

		if errInGetUrl != nil {
			retries -= 1
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	return response, errInGetUrl
}

func IsValidRegistryChartConfiguration(ociRegistry *sql.DockerArtifactStore) bool {
	if ociRegistry.OCIRegistryConfig == nil ||
		len(ociRegistry.OCIRegistryConfig) != 1 ||
		ociRegistry.OCIRegistryConfig[0].RepositoryType != sql.OCI_REGISRTY_REPO_TYPE_CHART ||
		ociRegistry.OCIRegistryConfig[0].RepositoryAction == sql.STORAGE_ACTION_TYPE_PUSH {
		return false
	}
	return true
}
