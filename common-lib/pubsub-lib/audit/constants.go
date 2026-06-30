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

// Package audit holds the shared audit-log contract: the structural
// MODULE / RESOURCE / ACTION vocabulary used to tag routes and the NATS
// message struct published by the orchestrator and consumed by the
// audit-log service.
package audit

// AuditModule is the top-level functional area an audited action belongs to.
// Values are stable, snake_case identifiers safe to persist and query on.
type AuditModule string

func (m AuditModule) ToString() string { return string(m) }

const (
	ModuleAppManagement             AuditModule = "application_management"
	ModuleInfrastructureManagement  AuditModule = "infrastructure_management"
	ModuleCostVisibility            AuditModule = "cost_visibility"
	ModuleSoftwareReleaseManagement AuditModule = "software_release_management"
	ModuleGlobalConfiguration       AuditModule = "global_configuration"
	ModuleUserManagement            AuditModule = "user_management"
)

// AuditResource is the kind of entity an audited action operates on.
// Seed the common ones here; extend as more routes are tagged.
type AuditResource string

func (r AuditResource) ToString() string { return string(r) }

const (
	ResourceApplication        AuditResource = "application"
	ResourceHelmApp            AuditResource = "helm_app"
	ResourceJob                AuditResource = "job"
	ResourceEnvironment        AuditResource = "environment"
	ResourceCluster            AuditResource = "cluster"
	ResourceCdPipeline         AuditResource = "cd_pipeline"
	ResourceCiPipeline         AuditResource = "ci_pipeline"
	ResourceWorkflow           AuditResource = "workflow"
	ResourceDeploymentTemplate AuditResource = "deployment_template"
	ResourceConfigMap          AuditResource = "config_map"
	ResourceSecret             AuditResource = "secret"
	ResourceGitMaterial        AuditResource = "git_material"
	ResourceUser               AuditResource = "user"
	ResourcePermissionGroup    AuditResource = "permission_group"
	ResourceUserGroup          AuditResource = "user_group"
	ResourceApiToken           AuditResource = "api_token"
	ResourceTeam               AuditResource = "team"
	// global configuration resources
	ResourceGitOpsConfig   AuditResource = "gitops_config"
	ResourceDockerRegistry AuditResource = "docker_registry"
	ResourceGitProvider    AuditResource = "git_provider"
	ResourceGitHost        AuditResource = "git_host"
	ResourceChartRepo      AuditResource = "chart_repo"
	ResourceSSOConfig      AuditResource = "sso_config"
	ResourceNotification   AuditResource = "notification_config"
	ResourceAuthorization  AuditResource = "authorization_config"
)

// AuditAction is the verb performed on the resource. It populates the
// AuditLogEvent.Type field; the human-readable sentence goes in Action.
type AuditAction string

func (a AuditAction) ToString() string { return string(a) }

const (
	ActionCreate      AuditAction = "create"
	ActionUpdate      AuditAction = "update"
	ActionDelete      AuditAction = "delete"
	ActionTrigger     AuditAction = "trigger"
	ActionDeploy      AuditAction = "deploy"
	ActionHibernate   AuditAction = "hibernate"
	ActionUnHibernate AuditAction = "unhibernate"
	ActionRollback    AuditAction = "rollback"
	ActionSync        AuditAction = "sync"
	ActionApprove     AuditAction = "approve"
	ActionClone       AuditAction = "clone"
	ActionAssign      AuditAction = "assign"
	ActionGet         AuditAction = "get"
)

// Enrichment entity keys used in AuditLogEvent.EnrichmentContext. These name
// the entity whose identifier the audit-log service resolves into a full
// record (e.g. "app" -> app_id -> application name/team/etc).
const (
	EntityApp         = "app"
	EntityUser        = "user"
	EntityEnvironment = "environment"
	EntityCluster     = "cluster"
	EntityTeam        = "team"
)
