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
	"errors"
)

var (
	// ErrorCacheMissReleaseNotFound is returned when a release is not found in the cache
	ErrorCacheMissReleaseNotFound = errors.New("release not found in cache")
	// InformerAlreadyExistError is returned when an informer already exists
	InformerAlreadyExistError = errors.New(INFORMER_ALREADY_EXIST_MESSAGE)
)

const (
	HELM_RELEASE_SECRET_TYPE       = "helm.sh/release.v1"
	INFORMER_ALREADY_EXIST_MESSAGE = "INFORMER_ALREADY_EXIST"
)
