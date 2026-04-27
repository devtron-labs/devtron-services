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
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/devtron-labs/git-sensor/bean"
	internalotel "github.com/devtron-labs/git-sensor/internals/otel"
	"github.com/devtron-labs/git-sensor/internals/sql"
	"github.com/devtron-labs/git-sensor/internals/tracing"
	"github.com/devtron-labs/git-sensor/pkg/git"
)

// tracedRepoManager is a decorator that wraps every RepoManager call with an
// OTel span.  It is wire-injected instead of the plain RepoManagerImpl when
// OTel is enabled.  Business logic is untouched — all tracing lives here.
//
// Methods that receive a git.GitContext already carry the parent span context
// (set by the gRPC handler via BuildGitContext(ctx)).  Methods without a
// GitContext create root spans; they are still visible in SigNoz as standalone
// operations with duration and error data.
type tracedRepoManager struct {
	inner RepoManager
}

// NewTracedRepoManager returns a traced wrapper when OTel is enabled,
// otherwise it returns inner unchanged.  Wire injects *RepoManagerImpl so
// the concrete type is resolved without a circular interface dependency.
func NewTracedRepoManager(inner *RepoManagerImpl) RepoManager {
	if !internalotel.Enabled {
		return inner
	}
	return &tracedRepoManager{inner: inner}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func startWithCtx(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	return tracing.Start(ctx, name, opts...)
}

// ── methods with GitContext (parent span flows through gitCtx.Context) ────────

func (t *tracedRepoManager) GetCommitMetadata(gitCtx git.GitContext, pipelineMaterialId int, gitHash string) (*git.GitCommitBase, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetCommitMetadata",
		attribute.Int("material.id", pipelineMaterialId),
		attribute.String("git.hash", gitHash))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetCommitMetadata(gitCtx, pipelineMaterialId, gitHash)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetLatestCommitForBranch(gitCtx git.GitContext, pipelineMaterialId int, branchName string) (*git.GitCommitBase, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetLatestCommitForBranch",
		attribute.Int("material.id", pipelineMaterialId),
		attribute.String("branch", branchName))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetLatestCommitForBranch(gitCtx, pipelineMaterialId, branchName)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetCommitMetadataForPipelineMaterial(gitCtx git.GitContext, pipelineMaterialId int, gitHash string) (*git.GitCommitBase, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetCommitMetadataForPipelineMaterial",
		attribute.Int("material.id", pipelineMaterialId))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetCommitMetadataForPipelineMaterial(gitCtx, pipelineMaterialId, gitHash)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) AddRepo(gitCtx git.GitContext, material []*sql.GitMaterial) ([]*sql.GitMaterial, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.AddRepo")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.AddRepo(gitCtx, material)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) UpdateRepo(gitCtx git.GitContext, material *sql.GitMaterial) (*sql.GitMaterial, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.UpdateRepo")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.UpdateRepo(gitCtx, material)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) SavePipelineMaterial(gitCtx git.GitContext, material []*sql.CiPipelineMaterial) ([]*sql.CiPipelineMaterial, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.SavePipelineMaterial")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.SavePipelineMaterial(gitCtx, material)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) ReloadAllRepo(gitCtx git.GitContext, req *bean.ReloadAllMaterialQuery) error {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.ReloadAllRepo")
	defer span.End()
	gitCtx.Context = ctx
	err := t.inner.ReloadAllRepo(gitCtx, req)
	tracing.End(span, err)
	return err
}

func (t *tracedRepoManager) ResetRepo(gitCtx git.GitContext, materialId int) error {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.ResetRepo",
		attribute.Int("material.id", materialId))
	defer span.End()
	gitCtx.Context = ctx
	err := t.inner.ResetRepo(gitCtx, materialId)
	tracing.End(span, err)
	return err
}

func (t *tracedRepoManager) GetReleaseChanges(gitCtx git.GitContext, request *ReleaseChangesRequest) (*git.GitChanges, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetReleaseChanges")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetReleaseChanges(gitCtx, request)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetCommitInfoForTag(gitCtx git.GitContext, request *git.CommitMetadataRequest) (*git.GitCommitBase, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetCommitInfoForTag")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetCommitInfoForTag(gitCtx, request)
	tracing.End(span, err)
	return result, err
}

// ── methods now with GitContext (context threaded from gRPC/REST handler) ─────

func (t *tracedRepoManager) GetHeadForPipelineMaterials(gitCtx git.GitContext, ids []int) ([]*git.CiPipelineMaterialBean, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetHeadForPipelineMaterials")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetHeadForPipelineMaterials(gitCtx, ids)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) FetchChanges(gitCtx git.GitContext, pipelineMaterialId int, from string, to string, count int, showAll bool) (*git.MaterialChangeResp, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.FetchChanges",
		attribute.Int("material.id", pipelineMaterialId))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.FetchChanges(gitCtx, pipelineMaterialId, from, to, count, showAll)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) SaveGitProvider(gitCtx git.GitContext, provider *sql.GitProvider) (*sql.GitProvider, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.SaveGitProvider")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.SaveGitProvider(gitCtx, provider)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) RefreshGitMaterial(gitCtx git.GitContext, req *git.RefreshGitMaterialRequest) (*git.RefreshGitMaterialResponse, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.RefreshGitMaterial")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.RefreshGitMaterial(gitCtx, req)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetWebhookAndCiDataById(gitCtx git.GitContext, id int, ciPipelineMaterialId int) (*git.WebhookAndCiData, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetWebhookAndCiDataById",
		attribute.Int("webhook.id", id))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetWebhookAndCiDataById(gitCtx, id, ciPipelineMaterialId)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetAllWebhookEventConfigForHost(gitCtx git.GitContext, req *git.WebhookEventConfigRequest) ([]*git.WebhookEventConfig, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetAllWebhookEventConfigForHost")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetAllWebhookEventConfigForHost(gitCtx, req)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetWebhookEventConfig(gitCtx git.GitContext, eventId int) (*git.WebhookEventConfig, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetWebhookEventConfig",
		attribute.Int("event.id", eventId))
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetWebhookEventConfig(gitCtx, eventId)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetWebhookPayloadDataForPipelineMaterialId(gitCtx git.GitContext, request *git.WebhookPayloadDataRequest) (*git.WebhookPayloadDataResponse, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetWebhookPayloadData")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetWebhookPayloadDataForPipelineMaterialId(gitCtx, request)
	tracing.End(span, err)
	return result, err
}

func (t *tracedRepoManager) GetWebhookPayloadFilterDataForPipelineMaterialId(gitCtx git.GitContext, request *git.WebhookPayloadFilterDataRequest) (*git.WebhookPayloadFilterDataResponse, error) {
	ctx, span := startWithCtx(gitCtx.Context, "RepoManager.GetWebhookPayloadFilterData")
	defer span.End()
	gitCtx.Context = ctx
	result, err := t.inner.GetWebhookPayloadFilterDataForPipelineMaterialId(gitCtx, request)
	tracing.End(span, err)
	return result, err
}
