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

package git

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/constants"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/git-sensor/internals"
	"github.com/devtron-labs/git-sensor/internals/middleware"
	"github.com/devtron-labs/git-sensor/internals/sql"
	util2 "github.com/devtron-labs/git-sensor/util"
	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"runtime/debug"
	"strings"
	"time"
)

type GitWatcherImpl struct {
	repositoryManager            RepositoryManager
	materialRepo                 sql.MaterialRepository
	cron                         *cron.Cron
	logger                       *zap.SugaredLogger
	ciPipelineMaterialRepository sql.CiPipelineMaterialRepository
	pubSubClient                 *pubsub.PubSubClientServiceImpl
	locker                       *internals.RepositoryLocker
	pollConfig                   *PollConfig
	webhookHandler               WebhookHandler
	configuration                *internals.Configuration
	gitManager                   GitManager
}

const PANIC = "panic"

type GitWatcher interface {
	PollAndUpdateGitMaterial(material *sql.GitMaterial) (*sql.GitMaterial, error)
}

type PollConfig struct {
	PollDuration int `env:"POLL_DURATION" envDefault:"2"`
	PollWorker   int `env:"POLL_WORKER" envDefault:"5"`
}

func NewGitWatcherImpl(repositoryManager RepositoryManager,
	materialRepo sql.MaterialRepository,
	logger *zap.SugaredLogger,
	ciPipelineMaterialRepository sql.CiPipelineMaterialRepository,
	locker *internals.RepositoryLocker,
	pubSubClient *pubsub.PubSubClientServiceImpl, webhookHandler WebhookHandler, configuration *internals.Configuration,
	gitmanager GitManager,
) (*GitWatcherImpl, error) {

	cfg := &PollConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	cronLogger := &CronLoggerImpl{logger: logger}
	cron := cron.New(
		cron.WithChain(
			cron.SkipIfStillRunning(cronLogger),
			cron.Recover(cronLogger)))
	cron.Start()
	watcher := &GitWatcherImpl{
		repositoryManager:            repositoryManager,
		cron:                         cron,
		logger:                       logger,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		materialRepo:                 materialRepo,
		locker:                       locker,
		pubSubClient:                 pubSubClient,
		pollConfig:                   cfg,
		webhookHandler:               webhookHandler,
		configuration:                configuration,
		gitManager:                   gitmanager,
	}

	logger.Info()
	_, err = cron.AddFunc(fmt.Sprintf("@every %dm", cfg.PollDuration), watcher.Watch)
	if err != nil {
		fmt.Println("error in starting cron")
		return nil, err
	}

	// err = watcher.SubscribePull()
	watcher.SubscribeWebhookEvent()
	return watcher, err
}
func (impl GitWatcherImpl) StopCron() {
	impl.cron.Stop()
}

func (impl GitWatcherImpl) Watch() {
	watchID := uuid.New().String()
	impl.logger.Infow("starting git watch thread", "watchID", watchID)
	materials, err := impl.materialRepo.FindActive()
	if err != nil {
		impl.logger.Error("error in fetching watchlist", "err", err, "watchID", watchID)
		return
	}
	// impl.Publish(materials)
	middleware.ActiveGitRepoCount.WithLabelValues().Set(float64(len(materials)))
	impl.RunOnWorker(materials)
	impl.logger.Infow("stop git watch thread", "watchID", watchID)
}

func (impl *GitWatcherImpl) RunOnWorker(materials []*sql.GitMaterial) {
	wp := workerpool.New(impl.pollConfig.PollWorker)

	handlePanic := func() {
		if err := recover(); err != nil {
			impl.logger.Error(constants.PanicLogIdentifier, "recovered from panic", "panic", err, "stack", string(debug.Stack()))

		}
	}

	for _, material := range materials {
		if len(material.CiPipelineMaterials) == 0 {
			impl.logger.Infow("no ci pipeline, skipping", "id", material.Id, "url", material.Url)
			continue
		}
		materialMsg := &sql.GitMaterial{Id: material.Id, Url: material.Url}
		wp.Submit(func() {
			defer handlePanic()
			_, err := impl.pollAndUpdateGitMaterial(materialMsg)
			if err != nil {
				impl.logger.Errorw("error in polling git material", "material", materialMsg, "err", err)
			}
		})
	}
	wp.StopWait()
}

func (impl GitWatcherImpl) PollAndUpdateGitMaterial(material *sql.GitMaterial) (*sql.GitMaterial, error) {
	// tmp expose remove in future
	return impl.pollAndUpdateGitMaterial(material)
}

func (impl GitWatcherImpl) pollAndUpdateGitMaterial(materialReq *sql.GitMaterial) (*sql.GitMaterial, error) {
	repoLock := impl.locker.LeaseLocker(materialReq.Id)
	repoLock.Mutex.Lock()
	defer func() {
		repoLock.Mutex.Unlock()
		impl.locker.ReturnLocker(materialReq.Id)
	}()
	material, err := impl.materialRepo.FindById(materialReq.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching material ", "material", materialReq, "err", err)
		return nil, err
	}
	errMsg, err := impl.pollGitMaterialAndNotify(material)
	material.LastFetchTime = time.Now()
	material.FetchStatus = err == nil
	if err != nil {
		material.LastFetchErrorCount = material.LastFetchErrorCount + 1
		material.FetchErrorMessage = util2.BuildDisplayErrorMessage(errMsg, err)
	} else {
		material.LastFetchErrorCount = 0
		material.FetchErrorMessage = ""
	}
	err = impl.materialRepo.Update(material)
	if err != nil {
		impl.logger.Errorw("error in updating fetch status", "material", material, "err", err)
	}
	return material, err
}

// Helper function to handle SSH key creation and retry fetching material
func (impl GitWatcherImpl) handleSshKeyCreationAndRetry(gitCtx GitContext, material *sql.GitMaterial, location string, gitProvider *sql.GitProvider) (updated bool, repo *GitRepository, errMsg string, err error) {
	if strings.Contains(material.CheckoutLocation, "/.git") {
		location, _, _, err = impl.repositoryManager.GetCheckoutLocationFromGitUrl(material, gitCtx.CloningMode)
		if err != nil {
			impl.logger.Errorw("error in getting clone location ", "material", material, "errMsg", errMsg, "err", err)
			return false, nil, errMsg, err
		}
	}
	_, errMsg, err = impl.repositoryManager.CreateSshFileIfNotExistsAndConfigureSshCommand(gitCtx, location, gitProvider.Id, gitProvider.SshPrivateKey)
	if err != nil {
		impl.logger.Errorw("error in creating/configuring ssh private key on disk ", "repo", material.Url, "gitProviderId", gitProvider.Id, "errMsg", errMsg, "err", err)
		return false, nil, errMsg, err
	} else {
		impl.logger.Info("Retrying fetching for", "repo", material.Url)
		updated, repo, errMsg, err = impl.FetchAndUpdateMaterial(gitCtx, material, location)
		if err != nil {
			impl.logger.Errorw("error in fetching material details in retry", "repo", material.Url, "err", err)
			errorMessage := util2.BuildDisplayErrorMessage(errMsg, err)
			material.FetchStatus = false
			material.FetchErrorMessage = errorMessage
			return false, nil, errMsg, err
		}
	}
	return updated, repo, errMsg, err
}

func (impl GitWatcherImpl) pollGitMaterialAndNotify(material *sql.GitMaterial) (string, error) {
	gitProvider := material.GitProvider
	userName, password, err := GetUserNamePassword(gitProvider)
	location := material.CheckoutLocation
	if err != nil {
		impl.logger.Errorw("error in determining location", "url", material.Url, "err", err)
		return "", err
	}
	gitCtx := BuildGitContext(context.Background()).
		WithCredentials(userName, password).
		WithTLSData(gitProvider.CaCert, gitProvider.TlsKey, gitProvider.TlsCert, material.GitProvider.EnableTLSVerification)

	updated, repo, errMsg, err := impl.FetchAndUpdateMaterial(gitCtx, material, location)
	if err != nil {
		impl.logger.Errorw("error in fetching material details ", "repo", material.Url, "errMsg", errMsg, "err", err)
		// there might be the case if ssh private key gets flush from disk, so creating and single retrying in this case
		// Retry mechanism for SSH-based authentication
		if gitProvider.AuthMode == sql.AUTH_MODE_SSH && !CheckIfSshPrivateKeyExists(gitProvider.Id) {
			updated, repo, errMsg, err = impl.handleSshKeyCreationAndRetry(gitCtx, material, location, gitProvider)
			if err != nil {
				impl.logger.Errorw("error in fetching material details in retry", "repo", material.Url, "materialId", material.Id, "errMsg", errMsg, "err", err)
				return errMsg, err
			}
		} else {
			// Log and update database if retry not possible or SSH key already exists
			impl.logger.Errorw("error in fetching material details in retry", "repo", material.Url, "materialId", material.Id, "err", err, "errMsg", errMsg)
			errorMessage := util2.BuildDisplayErrorMessage(errMsg, err)
			material.FetchStatus = false
			material.FetchErrorMessage = errorMessage
			return errMsg, err
		}
	}
	if !updated {
		impl.logger.Debugw("no new commit found but fetch success", "url", material.Url, "fetchStatus", material.FetchStatus)
		return "", nil
	}
	materials, err := impl.ciPipelineMaterialRepository.FindByGitMaterialId(material.Id)
	if err != nil {
		impl.logger.Errorw("error in calculating head", "err", err, "url", material.Url)
		return "", err
	}
	var updatedMaterials []*CiPipelineMaterialBean
	var updatedMaterialsModel []*sql.CiPipelineMaterial
	var erroredMaterialsModels []*sql.CiPipelineMaterial
	checkoutLocation := material.CheckoutLocation
	for _, material := range materials {
		if material.Type != sql.SOURCE_TYPE_BRANCH_FIXED {
			continue
		}
		impl.logger.Debugw("Running changesBySinceRepository for material - ", "materialId", material.Id)
		impl.logger.Debugw("---------------------------------------------------------- ")
		// parse env variables here, then search for the count field and pass here.
		lastSeenHash := ""
		if len(material.LastSeenHash) > 0 {
			// this might misbehave is the hash stored in table is corrupted somehow
			lastSeenHash = material.LastSeenHash
		}
		fetchCount := impl.configuration.GitHistoryCount
		commits, errMsg, err := impl.repositoryManager.ChangesSinceByRepository(gitCtx, repo, material.Value, lastSeenHash, "", fetchCount, checkoutLocation, false)
		impl.logger.Debugw("Got changesBySinceRepository for material ", "commits", commits, "errMsg", errMsg, "err", err)
		if err != nil {
			impl.logger.Errorw("error in fetching ChangesSinceByRepository", "err", err, "errMsg", errMsg, "Value", material.Value)
			material.Errored = true
			material.ErrorMsg = util2.BuildDisplayErrorMessage(errMsg, err)

			erroredMaterialsModels = append(erroredMaterialsModels, material)
		} else if len(commits) > 0 {
			latestCommit := commits[0]
			if latestCommit.GetCommit().Commit != material.LastSeenHash {

				commitsTotal, err := AppendOldCommitsFromHistory(commits, material.CommitHistory, fetchCount)
				if err != nil {
					impl.logger.Errorw("error in appending history to new commits", "material", material.GitMaterialId, "err", err)
				}

				// new commit found
				mb := &CiPipelineMaterialBean{
					Id:            material.Id,
					Value:         material.Value,
					GitMaterialId: material.GitMaterialId,
					Type:          material.Type,
					Active:        material.Active,
					GitCommit:     latestCommit,
				}
				updatedMaterials = append(updatedMaterials, mb)

				material.LastSeenHash = latestCommit.Commit
				material.CommitAuthor = latestCommit.Author
				material.CommitDate = latestCommit.Date
				material.CommitMessage = latestCommit.Message
				commitJson, _ := json.Marshal(commitsTotal)
				material.CommitHistory = string(commitJson)
				material.Errored = false
				material.ErrorMsg = ""
				updatedMaterialsModel = append(updatedMaterialsModel, material)
			}
			middleware.GitMaterialUpdateCounter.WithLabelValues().Inc()
		} else if len(commits) == 0 {
			// no new commit found, this is the case when no new commits are found but all git operations are working fine so update the error save in pipeline material to false
			material.Errored = false
			material.ErrorMsg = ""
			updatedMaterialsModel = append(updatedMaterialsModel, material)
		}
	}
	if len(erroredMaterialsModels) > 0 {
		err = impl.ciPipelineMaterialRepository.Update(erroredMaterialsModels)
		if err != nil {
			impl.logger.Errorw("error in update db ", "url", material.Url, "update", erroredMaterialsModels)
		}
	}
	if len(updatedMaterialsModel) > 0 {
		err = impl.ciPipelineMaterialRepository.Update(updatedMaterialsModel)
		if err != nil {
			impl.logger.Errorw("error in update db ", "url", material.Url, "update", updatedMaterialsModel)
		} else {
			err = impl.NotifyForMaterialUpdate(updatedMaterials, material)
			if err != nil {
				impl.logger.Errorw("error in sending notification for materials", "url", material.Url, "update", updatedMaterialsModel)
			}
		}
	}
	return "", nil
}

func (impl GitWatcherImpl) FetchAndUpdateMaterial(gitCtx GitContext, material *sql.GitMaterial, location string) (bool, *GitRepository, string, error) {
	updated, repo, errMsg, err := impl.repositoryManager.Fetch(gitCtx, material.Url, location)
	if err == nil {
		material.CheckoutLocation = location
		material.FetchStatus = true
	}
	return updated, repo, errMsg, err
}

func (impl GitWatcherImpl) NotifyForMaterialUpdate(materials []*CiPipelineMaterialBean, gitMaterial *sql.GitMaterial) error {

	impl.logger.Warnw("material notification", "materials", materials)
	for _, material := range materials {
		excluded := impl.gitManager.PathMatcher(material.GitCommit.FileStats, gitMaterial)
		if excluded {
			impl.logger.Infow("skip this auto trigger", "exclude", excluded)
			continue
		}
		mb, err := json.Marshal(material)
		if err != nil {
			impl.logger.Errorw("err in json marshaling", "err", err)
			continue
		}
		err = impl.pubSubClient.Publish(pubsub.NEW_CI_MATERIAL_TOPIC, string(mb))
		if err != nil {
			impl.logger.Errorw("error in publishing material modification msg ", "material", material)
		}

	}
	return nil
}

func (impl GitWatcherImpl) SubscribeWebhookEvent() error {
	callback := func(msg *model.PubSubMsg) {
		impl.logger.Debugw("received msg", "msg", msg)
		// msg.Ack() //ack immediate if lost next min it would get a new message
		webhookEvent := &WebhookEvent{}
		err := json.Unmarshal([]byte(msg.Data), webhookEvent)
		if err != nil {
			impl.logger.Infow("err in reading msg", "err", err)
			return
		}
		impl.webhookHandler.HandleWebhookEvent(webhookEvent)
	}

	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		webhookEvent := &WebhookEvent{}
		err := json.Unmarshal([]byte(msg.Data), &webhookEvent)
		if err != nil {
			return "error while unmarshalling WebhookEvent object", []interface{}{"err", err, "msg", msg.Data}
		}
		return "got message for WebhookEvent stage completion", []interface{}{"eventType", webhookEvent.EventType, "gitHostId", webhookEvent.GitHostId, "payloadId", webhookEvent.PayloadId}
	}
	err := impl.pubSubClient.Subscribe(pubsub.WEBHOOK_EVENT_TOPIC, callback, loggerFunc)
	return err
}

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	if msg == PANIC {
		metrics.IncPanicRecoveryCount("cron", "", "", "")
	}
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(msg, keysAndValues...)
}
