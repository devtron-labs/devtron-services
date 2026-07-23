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

package utils

import (
	"encoding/json"
	"strings"
)

// SecretDataMaskPlaceholder is written in place of any sensitive value that is masked out of a payload.
const SecretDataMaskPlaceholder = "********"

// unparseablePayloadPlaceholder is logged when a payload cannot be parsed as JSON. We fail closed here
// (redact the whole body) instead of returning the raw bytes, so an unexpected content type can never
// leak a secret into logs.
const unparseablePayloadPlaceholder = "[REDACTED: unparseable payload]"

// sensitivePayloadKeys are the (lower-cased) JSON field-name fragments whose values must never be logged.
// Matching is case-insensitive and substring based, so "password" also covers "confirmPassword",
// "sshPassword", etc., and "token" covers "accessToken", "dockerToken", and so on.
var sensitivePayloadKeys = []string{
	"password",
	"passphrase",
	"secret",
	"token",
	"credential",
	"privatekey",
	"apikey",
	"accesskey",
	"authkey",
}

// isSensitivePayloadKey reports whether a JSON key names a sensitive value that must be masked.
func isSensitivePayloadKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, sensitiveKey := range sensitivePayloadKeys {
		if strings.Contains(lowerKey, sensitiveKey) {
			return true
		}
	}
	return false
}

// MaskSensitiveDataInJsonPayload returns a JSON string equivalent to payload with every sensitive value
// (see sensitivePayloadKeys) replaced by SecretDataMaskPlaceholder. It is meant for the generic request
// logging / audit path, where only the raw body bytes are available and the concrete DTO type is unknown.
//
// It fails closed: an empty payload returns "", and a payload that cannot be parsed as JSON returns
// unparseablePayloadPlaceholder rather than the raw bytes, so a secret can never leak on the error path.
func MaskSensitiveDataInJsonPayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	var parsed interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return unparseablePayloadPlaceholder
	}
	maskSensitiveValues(parsed)
	masked, err := json.Marshal(parsed)
	if err != nil {
		return unparseablePayloadPlaceholder
	}
	return string(masked)
}

// maskSensitiveValues walks a decoded JSON value in place, masking any value whose key is sensitive and
// recursing into nested objects and arrays.
func maskSensitiveValues(value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, val := range typed {
			if isSensitivePayloadKey(key) {
				typed[key] = SecretDataMaskPlaceholder
			} else {
				maskSensitiveValues(val)
			}
		}
	case []interface{}:
		for _, item := range typed {
			maskSensitiveValues(item)
		}
	}
}
