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

package sql

import (
	"time"

	"context"

	pg "github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type AppRelease struct {
	tableName             struct{}      `pg:"app_release"`
	Id                    int           `pg:"id,pk"`
	AppId                 int           `pg:"app_id,notnull,use_zero"`                   //orchestrator appId
	EnvironmentId         int           `pg:"environment_id,notnull,use_zero"`           //orchestrator env id
	CiArtifactId          int           `pg:"ci_artifact_id,notnull,use_zero"`           //orchestrator ciAretefactId  used for identifying rollback (appId,environmentId, ciArtifactId)
	ReleaseId             int           `pg:"release_id,notnull,use_zero"`               // orchestrator release counter
	PipelineOverrideId    int           `pg:"pipeline_override_id,notnull,use_zero"`     //pipeline override id orchestrator
	ChangeSizeLineAdded   int           `pg:"change_size_line_added,notnull,use_zero"`   //total lines added in this release
	ChangeSizeLineDeleted int           `pg:"change_size_line_deleted,notnull,use_zero"` //total lines deleted during this release
	TriggerTime           time.Time     `pg:"trigger_time,notnull"`                      //deployment time
	ReleaseType           ReleaseType   `pg:"release_type,notnull,use_zero"`
	ReleaseStatus         ReleaseStatus `pg:"release_status,notnull,use_zero"`
	ProcessStage          ProcessStage  `pg:"process_status,notnull,use_zero"`
	CreatedTime           time.Time     `pg:"created_time,notnull"`
	UpdatedTime           time.Time     `pg:"updated_time,notnull"`
	LeadTime              *LeadTime
}

// --------------
type ReleaseStatus int

const (
	Success ReleaseStatus = iota
	Failure
)

func (releaseStatus ReleaseStatus) String() string {
	return [...]string{"Success", "Failure"}[releaseStatus]
}

// ----------------
type ReleaseType int

const (
	Unknown ReleaseType = iota
	RollForward
	RollBack
	Patch
)

func (releaseType ReleaseType) String() string {
	return [...]string{"Unknown", "RollForward", "RollBack", "Patch"}[releaseType]
}

// ------
type ProcessStage int

const (
	Init ProcessStage = iota
	ReleaseTypeDetermined
	LeadTimeFetch
)

var ctx = context.Background()

func (ProcessStage ProcessStage) String() string {
	return [...]string{"Init", "ReleaseTypeDetermined", "LeadTimeFetch"}[ProcessStage]
}

type AppReleaseRepository interface {
	Save(appRelease *AppRelease) (*AppRelease, error)
	Update(appRelease *AppRelease) (*AppRelease, error)
	CheckDuplicateRelease(appId, environmentId, ciArtifactId int) (bool, error)
	GetPreviousReleaseWithinTime(appId, environmentId int, within time.Time, currentAppReleaseId int) (*AppRelease, error)
	GetPreviousRelease(appId, environmentId int, appReleaseId int) (*AppRelease, error)
	GetReleaseBetween(appId, environmentId int, from time.Time, to time.Time) ([]AppRelease, error)
	CleanAppDataForEnvironment(appId, environmentId int) error
}
type AppReleaseRepositoryImpl struct {
	dbConnection               *pg.DB
	logger                     *zap.SugaredLogger
	leadTimeRepository         LeadTimeRepository
	pipelineMaterialRepository PipelineMaterialRepository
}

func NewAppReleaseRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger,
	leadTimeRepository LeadTimeRepository,
	pipelineMaterialRepository PipelineMaterialRepository) *AppReleaseRepositoryImpl {
	return &AppReleaseRepositoryImpl{logger: logger, dbConnection: dbConnection,
		leadTimeRepository:         leadTimeRepository,
		pipelineMaterialRepository: pipelineMaterialRepository}
}

func (impl *AppReleaseRepositoryImpl) Save(appRelease *AppRelease) (*AppRelease, error) {
	_, err := impl.dbConnection.Model(appRelease).Insert()
	return appRelease, err
}

func (impl *AppReleaseRepositoryImpl) Update(appRelease *AppRelease) (*AppRelease, error) {
	_, err := impl.dbConnection.Model(appRelease).WherePK().Update()
	return appRelease, err
}

func (impl *AppReleaseRepositoryImpl) CheckDuplicateRelease(appId, environmentId, ciArtifactId int) (bool, error) {
	var appRelease *AppRelease
	count, err := impl.dbConnection.
		Model(appRelease).
		Where("app_id = ?", appId).
		Where("environment_id =? ", environmentId).
		Where("ci_artifact_id =? ", ciArtifactId).
		Count()
	if err != nil {
		return false, err
	}
	return count > 1, nil
}

func (impl *AppReleaseRepositoryImpl) GetPreviousReleaseWithinTime(appId, environmentId int,
	within time.Time,
	currentAppReleaseId int) (*AppRelease, error) {
	appRelease := &AppRelease{}
	err := impl.dbConnection.
		Model(appRelease).
		Where("app_id = ?", appId).
		Where("environment_id =? ", environmentId).
		Where("trigger_time > ?", within).
		Where("id < ?", currentAppReleaseId).
		Last()
	return appRelease, err
}

func (impl *AppReleaseRepositoryImpl) GetPreviousRelease(appId, environmentId int,
	appReleaseId int) (*AppRelease, error) {
	appRelease := &AppRelease{}
	err := impl.dbConnection.
		Model(appRelease).
		Where("app_id = ?", appId).
		Where("environment_id =? ", environmentId).
		Where("id < ?", appReleaseId).
		Last()
	return appRelease, err
}

func (impl *AppReleaseRepositoryImpl) GetReleaseBetween(appId, environmentId int,
	from time.Time, //inclusive
	to time.Time, //inclusive
) ([]AppRelease, error) {
	var appReleases []AppRelease
	err := impl.dbConnection.
		Model(&appReleases).
		Where("app_id = ?", appId).
		Where("environment_id =? ", environmentId).
		Where("trigger_time >= ?", from).
		Where("trigger_time <= ?", to).
		Order("id desc").
		Select()
	return appReleases, err
}

func (impl *AppReleaseRepositoryImpl) cleanAppDataForEnvironment(appId, environmentId int, tx *pg.Tx) error {
	r, err := tx.Model((*AppRelease)(nil)).
		Where("app_id =?", appId).
		Where("environment_id =?", environmentId).
		Delete()
	if err != nil {
		impl.logger.Infow("AppRelease deleted for ", "app", appId, "env", environmentId, "count", r.RowsAffected())
		return nil
	} else {
		return err
	}
}

func (impl *AppReleaseRepositoryImpl) CleanAppDataForEnvironment(appId, environmentId int) error {
	err := impl.dbConnection.RunInTransaction(ctx, func(tx *pg.Tx) error {
		err := impl.leadTimeRepository.CleanAppDataForEnvironment(appId, environmentId, tx)
		if err != nil {
			impl.logger.Errorw("error in cleaning pipeline", "appId", appId, "environmentId", environmentId, "err", err)
			return err
		}
		err = impl.pipelineMaterialRepository.CleanAppDataForEnvironment(appId, environmentId, tx)
		if err != nil {
			impl.logger.Errorw("error in cleaning pipeline", "appId", appId, "environmentId", environmentId, "err", err)
			return err
		}
		err = impl.cleanAppDataForEnvironment(appId, environmentId, tx)
		if err != nil {
			impl.logger.Errorw("error in cleaning AppRelease", "appId", appId, "environmentId", environmentId, "err", err)
			return err
		}
		return nil
	})
	return err
}
