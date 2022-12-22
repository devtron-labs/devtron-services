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
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func ReadFromPublicUrlWithRetry(url string) ([]byte, error) {
	var (
		err      error
		response *http.Response
		retries  = 3
	)

	for retries > 0 {
		response, err = http.Get(url)
		if err != nil {
			retries -= 1
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if response != nil {
		defer response.Body.Close()
		statusCode := response.StatusCode
		if statusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("Error in getting content from url - %s. Status code : %s", url, strconv.Itoa(statusCode)))
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, err
}

func ReadFromPrivateUrlWithRetry(baseurl string, absoluteUrl string, username string, password string, allowInsecureConnection bool) ([]byte, error) {
	var (
		err, errInGetUrl error
		response         *bytes.Buffer
		retries          = 3
	)
	getters := getter.All(&cli.EnvSettings{})
	u, err := url.Parse(baseurl)
	if err != nil {
		return nil, errors.Errorf("invalid chart URL format: %s", baseurl)
	}
	client, err := getters.ByScheme(u.Scheme)
	if err != nil {
		return nil, errors.Errorf("could not find protocol handler for: %s", u.Scheme)
	}

	for retries > 0 {
		response, errInGetUrl = client.Get(absoluteUrl,
			getter.WithURL(baseurl),
			getter.WithInsecureSkipVerifyTLS(allowInsecureConnection),
			getter.WithBasicAuth(username, password),
		)

		if errInGetUrl != nil {
			retries -= 1
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if response != nil {
		resp, err := http.Get(absoluteUrl)
		defer resp.Body.Close()
		statusCode := resp.StatusCode
		if statusCode != http.StatusOK && errInGetUrl != nil {
			return nil, errors.New(fmt.Sprintf("Error in getting content from url - %s. Status code : %s", absoluteUrl, strconv.Itoa(statusCode)))
		}
		body, err := ioutil.ReadAll(response)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, err
}
