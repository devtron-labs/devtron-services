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

// Values are human-readable and surfaced directly on the UI, so they use spaces
// rather than underscores.
const (
	ModuleAppManagement             AuditModule = "Application Management"
	ModuleInfrastructureManagement  AuditModule = "Infrastructure Management"
	ModuleCostVisibility            AuditModule = "Cost Visibility"
	ModuleSoftwareReleaseManagement AuditModule = "Software Release Management"
	ModuleGlobalConfiguration       AuditModule = "Global Configuration"
	ModuleUserManagement            AuditModule = "User Management"
)

// AuditResource is the kind of entity an audited action operates on.
// Seed the common ones here; extend as more routes are tagged.
type AuditResource string

func (r AuditResource) ToString() string { return string(r) }

// Values are human-readable and surfaced directly on the UI, so they use spaces
// rather than underscores.
const (
	ResourceApplication        AuditResource = "Application"
	ResourceHelmApp            AuditResource = "Helm App"
	ResourceJob                AuditResource = "Job"
	ResourceEnvironment        AuditResource = "Environment"
	ResourceCluster            AuditResource = "Cluster"
	ResourceCdPipeline         AuditResource = "Cd Pipeline"
	ResourceCiPipeline         AuditResource = "Ci Pipeline"
	ResourceBuildConfig        AuditResource = "Build Config"
	ResourceWorkflow           AuditResource = "Workflow"
	ResourceDeploymentTemplate AuditResource = "Deployment Template"
	ResourceConfigMap          AuditResource = "Config Map"
	ResourceSecret             AuditResource = "Secret"
	ResourceGitMaterial        AuditResource = "Git Material"
	ResourcePod                AuditResource = "Pod"
	ResourceUser               AuditResource = "User"
	ResourcePermissionGroup    AuditResource = "Permission Group"
	ResourceUserGroup          AuditResource = "User Group"
	ResourceApiToken           AuditResource = "Api Token"
	ResourceTeam               AuditResource = "Project" // "team" is the internal name; the product/UI calls it a Project
	// global configuration resources
	ResourceGitOpsConfig   AuditResource = "Gitops Config"
	ResourceDockerRegistry AuditResource = "Docker Registry"
	ResourceGitProvider    AuditResource = "Git Provider"
	ResourceGitHost        AuditResource = "Git Host"
	ResourceChartRepo      AuditResource = "Chart Repo"
	ResourceSSOConfig      AuditResource = "Sso Config"
	ResourceNotification   AuditResource = "Notification Config"
	ResourceAuthorization  AuditResource = "Authorization Config"
	// infrastructure management resources (resource browser)
	ResourceK8sResource AuditResource = "Kubernetes Resource"
	ResourceNode        AuditResource = "Node"
	// application management resources
	ResourceDeploymentWindowProfile AuditResource = "Deployment Window Profile"
	// security & release-gating resources
	ResourceVulnerabilityPolicy     AuditResource = "Vulnerability Policy"
	ResourceArtifactPromotionPolicy AuditResource = "Artifact Promotion Policy"
	ResourceImageApproval           AuditResource = "Image Approval"
)

// AuditAction is the verb performed on the resource. It populates the
// AuditLogEvent.Type field; the human-readable sentence goes in Action.
type AuditAction string

func (a AuditAction) ToString() string { return string(a) }

const (
	ActionCreate      AuditAction = "Create"
	ActionUpdate      AuditAction = "Update"
	ActionDelete      AuditAction = "Delete"
	ActionTrigger     AuditAction = "Trigger"
	ActionDeploy      AuditAction = "Deploy"
	ActionHibernate   AuditAction = "Hibernate"
	ActionUnHibernate AuditAction = "Unhibernate"
	ActionRollback    AuditAction = "Rollback"
	ActionSync        AuditAction = "Sync"
	ActionApprove     AuditAction = "Approve"
	ActionClone       AuditAction = "Clone"
	ActionAssign      AuditAction = "Assign"
	ActionGet         AuditAction = "Get"
	ActionExec        AuditAction = "Exec"
	ActionApply       AuditAction = "Apply"
)

// Enrichment entity keys used in AuditLogEvent.EnrichmentContext. These name
// the entity whose identifier the audit-log service resolves into a full
// record (e.g. "app" -> app_id -> application name/team/etc).
const (
	EntityApp         = "App"
	EntityUser        = "User"
	EntityEnvironment = "Environment"
	EntityCluster     = "Cluster"
	EntityTeam        = "Team"
	EntityPipeline    = "Pipeline"
)
