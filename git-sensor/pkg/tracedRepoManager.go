/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package pkg

import (
	"go.opentelemetry.io/otel/attribute"

	"github.com/devtron-labs/git-sensor/bean"
	internalotel "github.com/devtron-labs/git-sensor/internals/otel"
	"github.com/devtron-labs/git-sensor/internals/sql"
	"github.com/devtron-labs/git-sensor/internals/tracing"
	"github.com/devtron-labs/git-sensor/pkg/git"
)

// tracedRepoManager is a decorator that wraps every RepoManager call with an
// OTel span.  Span names are derived automatically from the calling method via
// runtime.Caller — no hardcoded string literals to maintain.
//
// Wire-injected instead of the plain RepoManagerImpl when OTel is enabled.
// Business logic is untouched — all tracing lives here.
type tracedRepoManager struct {
	inner RepoManager
}

// NewTracedRepoManager returns a traced wrapper when OTel is enabled,
// otherwise it returns inner unchanged.
func NewTracedRepoManager(inner *RepoManagerImpl) RepoManager {
	if !internalotel.Enabled {
		return inner
	}
	return &tracedRepoManager{inner: inner}
}

// ── methods with GitContext ───────────────────────────────────────────────────

func (t *tracedRepoManager) GetCommitMetadata(gitCtx git.GitContext, pipelineMaterialId int, gitHash string) (result *git.GitCommitBase, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("material.id", pipelineMaterialId),
		attribute.String("git.hash", gitHash))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetCommitMetadata(gitCtx, pipelineMaterialId, gitHash)
	return
}

func (t *tracedRepoManager) GetLatestCommitForBranch(gitCtx git.GitContext, pipelineMaterialId int, branchName string) (result *git.GitCommitBase, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("material.id", pipelineMaterialId),
		attribute.String("branch", branchName))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetLatestCommitForBranch(gitCtx, pipelineMaterialId, branchName)
	return
}

func (t *tracedRepoManager) GetCommitMetadataForPipelineMaterial(gitCtx git.GitContext, pipelineMaterialId int, gitHash string) (result *git.GitCommitBase, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("material.id", pipelineMaterialId))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetCommitMetadataForPipelineMaterial(gitCtx, pipelineMaterialId, gitHash)
	return
}

func (t *tracedRepoManager) AddRepo(gitCtx git.GitContext, material []*sql.GitMaterial) (result []*sql.GitMaterial, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.AddRepo(gitCtx, material)
	return
}

func (t *tracedRepoManager) UpdateRepo(gitCtx git.GitContext, material *sql.GitMaterial) (result *sql.GitMaterial, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.UpdateRepo(gitCtx, material)
	return
}

func (t *tracedRepoManager) SavePipelineMaterial(gitCtx git.GitContext, material []*sql.CiPipelineMaterial) (result []*sql.CiPipelineMaterial, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.SavePipelineMaterial(gitCtx, material)
	return
}

func (t *tracedRepoManager) ReloadAllRepo(gitCtx git.GitContext, req *bean.ReloadAllMaterialQuery) (err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	err = t.inner.ReloadAllRepo(gitCtx, req)
	return
}

func (t *tracedRepoManager) ResetRepo(gitCtx git.GitContext, materialId int) (err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("material.id", materialId))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	err = t.inner.ResetRepo(gitCtx, materialId)
	return
}

func (t *tracedRepoManager) GetReleaseChanges(gitCtx git.GitContext, request *ReleaseChangesRequest) (result *git.GitChanges, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetReleaseChanges(gitCtx, request)
	return
}

func (t *tracedRepoManager) GetCommitInfoForTag(gitCtx git.GitContext, request *git.CommitMetadataRequest) (result *git.GitCommitBase, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetCommitInfoForTag(gitCtx, request)
	return
}

func (t *tracedRepoManager) GetHeadForPipelineMaterials(gitCtx git.GitContext, ids []int) (result []*git.CiPipelineMaterialBean, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetHeadForPipelineMaterials(gitCtx, ids)
	return
}

func (t *tracedRepoManager) FetchChanges(gitCtx git.GitContext, pipelineMaterialId int, from string, to string, count int, showAll bool) (result *git.MaterialChangeResp, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("material.id", pipelineMaterialId))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.FetchChanges(gitCtx, pipelineMaterialId, from, to, count, showAll)
	return
}

func (t *tracedRepoManager) SaveGitProvider(gitCtx git.GitContext, provider *sql.GitProvider) (result *sql.GitProvider, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.SaveGitProvider(gitCtx, provider)
	return
}

func (t *tracedRepoManager) RefreshGitMaterial(gitCtx git.GitContext, req *git.RefreshGitMaterialRequest) (result *git.RefreshGitMaterialResponse, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.RefreshGitMaterial(gitCtx, req)
	return
}

func (t *tracedRepoManager) GetWebhookAndCiDataById(gitCtx git.GitContext, id int, ciPipelineMaterialId int) (result *git.WebhookAndCiData, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("webhook.id", id))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetWebhookAndCiDataById(gitCtx, id, ciPipelineMaterialId)
	return
}

func (t *tracedRepoManager) GetAllWebhookEventConfigForHost(gitCtx git.GitContext, req *git.WebhookEventConfigRequest) (result []*git.WebhookEventConfig, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetAllWebhookEventConfigForHost(gitCtx, req)
	return
}

func (t *tracedRepoManager) GetWebhookEventConfig(gitCtx git.GitContext, eventId int) (result *git.WebhookEventConfig, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context,
		attribute.Int("event.id", eventId))
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetWebhookEventConfig(gitCtx, eventId)
	return
}

func (t *tracedRepoManager) GetWebhookPayloadDataForPipelineMaterialId(gitCtx git.GitContext, request *git.WebhookPayloadDataRequest) (result *git.WebhookPayloadDataResponse, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetWebhookPayloadDataForPipelineMaterialId(gitCtx, request)
	return
}

func (t *tracedRepoManager) GetWebhookPayloadFilterDataForPipelineMaterialId(gitCtx git.GitContext, request *git.WebhookPayloadFilterDataRequest) (result *git.WebhookPayloadFilterDataResponse, err error) {
	ctx, span := tracing.StartCaller(gitCtx.Context)
	defer func() { tracing.End(span, err) }()
	gitCtx.Context = ctx
	result, err = t.inner.GetWebhookPayloadFilterDataForPipelineMaterialId(gitCtx, request)
	return
}
