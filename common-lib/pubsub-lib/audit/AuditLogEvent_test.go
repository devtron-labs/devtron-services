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

package audit

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditLogEvent_MarshalShape(t *testing.T) {
	ts := time.Date(2026, 6, 29, 6, 7, 11, 0, time.UTC)
	event := NewAuditLogEvent("/orchestrator/app/edit", ts, ModuleAppManagement, ResourceApplication, ActionCreate).
		WithResourceName("dashboard").
		WithAction("Created application 'dashboard'").
		WithPayloadField("request_path", "/orchestrator/app/edit").
		WithPayloadField("http_method", "POST").
		WithEnrichment(EntityApp, "42").
		WithEnrichment(EntityUser, "7").
		WithEnrichment(EntityCluster, "") // empty identifier must be ignored

	out, err := event.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Round-trip into a generic map to assert the wire contract.
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["apiPath"] != "/orchestrator/app/edit" {
		t.Errorf("apiPath = %v", m["apiPath"])
	}
	if m["module"] != "application_management" {
		t.Errorf("module = %v", m["module"])
	}
	if m["type"] != "create" {
		t.Errorf("type = %v", m["type"])
	}
	if m["time"] != "2026-06-29T06:07:11+0000" {
		t.Errorf("time = %v", m["time"])
	}
	resource := m["resource"].(map[string]interface{})
	if resource["type"] != "application" || resource["name"] != "dashboard" {
		t.Errorf("resource = %v", resource)
	}
	enrich := m["enrichmentContext"].(map[string]interface{})
	if _, ok := enrich[EntityCluster]; ok {
		t.Errorf("empty-identifier entity must be dropped, got %v", enrich)
	}
	if enrich[EntityApp].(map[string]interface{})["identifier"] != "42" {
		t.Errorf("app identifier = %v", enrich[EntityApp])
	}

	// Round-trip back to a typed event.
	back, err := UnmarshalAuditLogEvent(out)
	if err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if back.Resource.Name != "dashboard" || back.EnrichmentContext[EntityUser].Identifier != "7" {
		t.Errorf("round-trip mismatch: %+v", back)
	}
}
