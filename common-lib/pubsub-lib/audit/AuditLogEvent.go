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
	"time"
)

// TimeFormat is the wire format for AuditLogEvent.Time (ISO-8601, numeric zone).
const TimeFormat = "2006-01-02T15:04:05-0700"

// RouteName builds the MODULE:RESOURCE:ACTION route name read back by the
// orchestrator's audit log service. Using this helper keeps route tags
// consistent with the typed Module/Resource/Action constants instead of
// hand-written strings.
func RouteName(module AuditModule, resource AuditResource, action AuditAction) string {
	return string(module) + ":" + string(resource) + ":" + string(action)
}

// ResourceRef identifies the concrete resource an audited action touched.
type ResourceRef struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// EnrichmentEntity carries the raw identifier (e.g. an app_id) that the
// audit-log service resolves into the full entity during enrichment.
type EnrichmentEntity struct {
	Identifier string `json:"identifier"`
}

// AuditLogEvent is the canonical NATS payload published on AUDIT_LOG_TOPIC.
// It is shared between the publisher (orchestrator) and the consumer
// (audit-log service) so both agree on the contract.
type AuditLogEvent struct {
	ApiPath           string                      `json:"apiPath"`
	Time              string                      `json:"time"`
	Resource          ResourceRef                 `json:"resource"`
	Module            string                      `json:"module"`
	Type              string                      `json:"type"`   // action verb, e.g. "create"
	Action            string                      `json:"action"` // human sentence, e.g. "Created application 'dashboard'"
	Payload           map[string]interface{}      `json:"payload,omitempty"`
	EnrichmentContext map[string]EnrichmentEntity `json:"enrichmentContext,omitempty"`
}

// NewAuditLogEvent constructs an event with the structural metadata parsed
// from a route name. eventTime is supplied by the caller (time.Now()) so this
// package stays free of implicit clock reads.
func NewAuditLogEvent(apiPath string, eventTime time.Time, module AuditModule, resourceType AuditResource, action AuditAction) *AuditLogEvent {
	return &AuditLogEvent{
		ApiPath:           apiPath,
		Time:              eventTime.Format(TimeFormat),
		Module:            module.ToString(),
		Type:              action.ToString(),
		Resource:          ResourceRef{Type: resourceType.ToString()},
		Payload:           make(map[string]interface{}),
		EnrichmentContext: make(map[string]EnrichmentEntity),
	}
}

// WithResourceName sets the human-facing name of the affected resource.
func (e *AuditLogEvent) WithResourceName(name string) *AuditLogEvent {
	e.Resource.Name = name
	return e
}

// WithAction sets the human-readable action sentence.
func (e *AuditLogEvent) WithAction(action string) *AuditLogEvent {
	e.Action = action
	return e
}

// WithPayloadField adds a single key/value to the payload bag.
func (e *AuditLogEvent) WithPayloadField(key string, value interface{}) *AuditLogEvent {
	if e.Payload == nil {
		e.Payload = make(map[string]interface{})
	}
	e.Payload[key] = value
	return e
}

// WithEnrichment records an identifier (e.g. app_id) under an entity key
// (e.g. "app") for the consumer to resolve. Empty identifiers are ignored so
// callers can pass through missing path vars without polluting the context.
func (e *AuditLogEvent) WithEnrichment(entity, identifier string) *AuditLogEvent {
	if identifier == "" {
		return e
	}
	if e.EnrichmentContext == nil {
		e.EnrichmentContext = make(map[string]EnrichmentEntity)
	}
	e.EnrichmentContext[entity] = EnrichmentEntity{Identifier: identifier}
	return e
}

// Marshal serializes the event to a JSON string for PubSubClientService.Publish.
func (e *AuditLogEvent) Marshal() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalAuditLogEvent parses a NATS payload back into an event (consumer side).
func UnmarshalAuditLogEvent(data string) (*AuditLogEvent, error) {
	event := &AuditLogEvent{}
	if err := json.Unmarshal([]byte(data), event); err != nil {
		return nil, err
	}
	return event, nil
}
