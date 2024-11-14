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

/*
 * grafeas.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package grafeas

// An entity that can have metadata. For example, a Docker image.
type V1beta1Resource struct {
	// Deprecated, do not use. Use uri instead.  The name of the resource. For example, the name of a Docker image - \"Debian\".
	Name string `json:"name,omitempty"`
	// Required. The unique URI of the resource. For example, `https://gcr.io/project/image@sha256:foo` for a Docker image.
	Uri string `json:"uri,omitempty"`
	// Deprecated, do not use. Use uri instead.  The hash of the resource content. For example, the Docker digest.
	ContentHash *ProvenanceHash `json:"content_hash,omitempty"`
}
