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

package error

import (
	customErr "github.com/devtron-labs/common-lib/k8sResource/errors"
)

// ConvertHelmErrorToInternalError converts known error message from helm to internal error and also maps it with proper grpc code
func ConvertHelmErrorToInternalError(err error) error {
	return customErr.ConvertHelmErrorToInternalError(err)
}
