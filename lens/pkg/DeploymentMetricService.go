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

package pkg

import (
	"fmt"
	"time"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/lens/internal/dto"
	"github.com/devtron-labs/lens/internal/sql"
	"go.uber.org/zap"
)

const (
	layout = "2006-01-02T15:04:05.000Z"
)

type DeploymentMetricService interface {
	GetDeploymentMetrics(request *dto.MetricRequest) (*dto.Metrics, error)
	GetBulkDeploymentMetrics(request *dto.BulkMetricRequest) (*dto.BulkMetricsResponse, error)
}

type DeploymentMetricServiceImpl struct {
	logger                     *zap.SugaredLogger
	appReleaseRepository       sql.AppReleaseRepository
	pipelineMaterialRepository sql.PipelineMaterialRepository
	leadTimeRepository         sql.LeadTimeRepository
}

func NewDeploymentMetricServiceImpl(
	logger *zap.SugaredLogger,
	appReleaseRepository sql.AppReleaseRepository,
	pipelineMaterialRepository sql.PipelineMaterialRepository,
	leadTimeRepository sql.LeadTimeRepository) *DeploymentMetricServiceImpl {
	return &DeploymentMetricServiceImpl{
		logger:                     logger,
		appReleaseRepository:       appReleaseRepository,
		pipelineMaterialRepository: pipelineMaterialRepository,
		leadTimeRepository:         leadTimeRepository,
	}
}

func (impl DeploymentMetricServiceImpl) GetDeploymentMetrics(request *dto.MetricRequest) (*dto.Metrics, error) {
	from, to, err := impl.parseDateRange(request.From, request.To)
	if err != nil {
		return nil, err
	}

	releases, err := impl.appReleaseRepository.GetReleaseBetween(request.AppId, request.EnvId, from, to)
	if err != nil {
		impl.logger.Errorw("error getting data from db ", "err", err)
		return nil, err
	}

	if len(releases) == 0 {
		return impl.createEmptyMetrics(), nil
	}
	var releaseIds []int
	for _, v := range releases {
		releaseIds = append(releaseIds, v.Id)
	}

	materials, err := impl.pipelineMaterialRepository.FindByAppReleaseIds(releaseIds)
	if err != nil {
		impl.logger.Errorw("error getting material from db ", "err", err)
		return nil, err
	}

	leadTimes, err := impl.leadTimeRepository.FindByIds(releaseIds)
	if err != nil {
		impl.logger.Errorw("error getting lead time from db ", "err", err)
		return nil, err
	}

	// Get previous release with bounds checking
	var lastRelease *sql.AppRelease
	if len(releases) > 0 {
		lastId := releases[len(releases)-1].Id
		lastRelease, err = impl.appReleaseRepository.GetPreviousRelease(request.AppId, request.EnvId, lastId)
		if err != nil && !utils.IsErrNoRows(err) {
			impl.logger.Errorw("error getting previous release from db ", "err", err)
			// Don't return error, just continue without previous release
		}
		if utils.IsErrNoRows(err) {
			lastRelease = nil
		}
	}

	return impl.populateMetrics(releases, materials, leadTimes, lastRelease)
}

func (impl DeploymentMetricServiceImpl) GetBulkDeploymentMetrics(request *dto.BulkMetricRequest) (*dto.BulkMetricsResponse, error) {
	if len(request.AppEnvPairs) == 0 {
		return &dto.BulkMetricsResponse{Results: []dto.AppEnvMetrics{}}, nil
	}

	response := &dto.BulkMetricsResponse{
		Results: make([]dto.AppEnvMetrics, len(request.AppEnvPairs)),
	}

	return impl.getBulkDeploymentMetricsWithBulkQueries(request, response)
}

func (impl DeploymentMetricServiceImpl) getBulkDeploymentMetricsWithBulkQueries(request *dto.BulkMetricRequest, response *dto.BulkMetricsResponse) (*dto.BulkMetricsResponse, error) {
	from, to, err := impl.parseDateRange(request.From, request.To)
	if err != nil {
		impl.logger.Errorw("error parsing date range", "from", request.From, "to", request.To, "err", err)
		return nil, err
	}

	// Step 1: Get all releases for all app-env pairs in one query
	allReleases, err := impl.appReleaseRepository.GetReleaseBetweenBulk(request.AppEnvPairs, from, to)
	if err != nil {
		impl.logger.Errorw("error getting bulk releases from db", "err", err)
		return nil, err
	}

	// Step 2: Group releases by app-env pair
	releasesByAppEnv := make(map[string][]sql.AppRelease)
	var allReleaseIds []int

	for _, release := range allReleases {
		key := impl.generateAppEnvKey(release.AppId, release.EnvironmentId)
		releasesByAppEnv[key] = append(releasesByAppEnv[key], release)
		allReleaseIds = append(allReleaseIds, release.Id)
	}

	// Step 3: Get all materials and lead times in bulk
	var allMaterials []*sql.PipelineMaterial
	var allLeadTimes []sql.LeadTime

	if len(allReleaseIds) > 0 {
		allMaterials, err = impl.pipelineMaterialRepository.FindByAppReleaseIds(allReleaseIds)
		if err != nil {
			impl.logger.Errorw("error getting bulk materials from db", "err", err)
			return nil, err
		}

		allLeadTimes, err = impl.leadTimeRepository.FindByIds(allReleaseIds)
		if err != nil {
			impl.logger.Errorw("error getting bulk lead times from db", "err", err)
			return nil, err
		}
	}

	// Step 4: Get previous releases for all app-env pairs
	previousReleases, err := impl.appReleaseRepository.GetPreviousReleasesBulk(allReleases)
	if err != nil {
		impl.logger.Errorw("error getting bulk previous releases from db", "err", err)
		return nil, err
	}

	// Step 5: Process each app-env pair
	for i, pair := range request.AppEnvPairs {
		key := impl.generateAppEnvKey(pair.AppId, pair.EnvId)
		releases := releasesByAppEnv[key]

		appEnvMetric := dto.AppEnvMetrics{
			AppId: pair.AppId,
			EnvId: pair.EnvId,
		}

		if len(releases) == 0 {
			appEnvMetric.Metrics = impl.createEmptyMetrics()
		} else {
			metrics, err := impl.processAppEnvMetrics(releases, allMaterials, allLeadTimes, previousReleases[key])
			if err != nil {
				appEnvMetric.Error = err.Error()
				impl.logger.Errorw("error populating metrics for app-env pair", "appId", pair.AppId, "envId", pair.EnvId, "err", err)
			} else {
				appEnvMetric.Metrics = metrics
			}
		}

		response.Results[i] = appEnvMetric
	}

	return response, nil
}

// parseDateRange parses from and to date strings
func (impl DeploymentMetricServiceImpl) parseDateRange(from, to string) (time.Time, time.Time, error) {
	fromTime, err := time.Parse(layout, from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	toTime, err := time.Parse(layout, to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return fromTime, toTime, nil
}

// generateAppEnvKey creates a consistent key for app-env pair mapping
func (impl DeploymentMetricServiceImpl) generateAppEnvKey(appId, envId int) string {
	return fmt.Sprintf("%d-%d", appId, envId)
}

// createEmptyMetrics creates an empty metrics response
func (impl DeploymentMetricServiceImpl) createEmptyMetrics() *dto.Metrics {
	return &dto.Metrics{Series: []*dto.Metric{}}
}

// processAppEnvMetrics processes metrics for a single app-env pair
func (impl DeploymentMetricServiceImpl) processAppEnvMetrics(releases []sql.AppRelease, allMaterials []*sql.PipelineMaterial, allLeadTimes []sql.LeadTime, previousRelease *sql.AppRelease) (*dto.Metrics, error) {
	releaseIds := make([]int, len(releases))
	for i, release := range releases {
		releaseIds[i] = release.Id
	}

	// Filter materials and lead times for this app-env pair
	materials := impl.filterMaterialsByReleaseIds(allMaterials, releaseIds)
	leadTimes := impl.filterLeadTimesByReleaseIds(allLeadTimes, releaseIds)

	return impl.populateMetrics(releases, materials, leadTimes, previousRelease)
}

// filterMaterialsByReleaseIds filters materials for specific release IDs
func (impl DeploymentMetricServiceImpl) filterMaterialsByReleaseIds(allMaterials []*sql.PipelineMaterial, releaseIds []int) []*sql.PipelineMaterial {
	if len(releaseIds) == 0 {
		return []*sql.PipelineMaterial{}
	}

	releaseIdSet := make(map[int]bool, len(releaseIds))
	for _, id := range releaseIds {
		releaseIdSet[id] = true
	}

	filtered := make([]*sql.PipelineMaterial, 0, len(releaseIds))
	for _, material := range allMaterials {
		if releaseIdSet[material.AppReleaseId] {
			filtered = append(filtered, material)
		}
	}
	return filtered
}

// filterLeadTimesByReleaseIds filters lead times for specific release IDs
func (impl DeploymentMetricServiceImpl) filterLeadTimesByReleaseIds(allLeadTimes []sql.LeadTime, releaseIds []int) []sql.LeadTime {
	if len(releaseIds) == 0 {
		return []sql.LeadTime{}
	}

	releaseIdSet := make(map[int]bool, len(releaseIds))
	for _, id := range releaseIds {
		releaseIdSet[id] = true
	}

	filtered := make([]sql.LeadTime, 0, len(releaseIds)) // Pre-allocate with estimated capacity
	for _, leadTime := range allLeadTimes {
		if releaseIdSet[leadTime.AppReleaseId] {
			filtered = append(filtered, leadTime)
		}
	}
	return filtered
}

func (impl DeploymentMetricServiceImpl) populateMetrics(appReleases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, lastRelease *sql.AppRelease) (*dto.Metrics, error) {
	releases := impl.transform(appReleases, materials, leadTimes)
	leadTimesCount := 0
	totalLeadTime := float64(0)
	for _, r := range releases {
		if r.LeadTime != float64(0) {
			totalLeadTime += r.LeadTime
			leadTimesCount++
		}
	}

	totalCycleTime := float64(0)
	cycleTimeCount := len(releases)
	for i := 0; i < len(releases)-1; i++ {
		releases[i].CycleTime = releases[i].ReleaseTime.Sub(releases[i+1].ReleaseTime).Minutes()
		totalCycleTime += releases[i].CycleTime
	}
	if lastRelease != nil {
		releases[len(releases)-1].CycleTime = releases[len(releases)-1].ReleaseTime.Sub(lastRelease.TriggerTime).Minutes()
		totalCycleTime += releases[len(releases)-1].CycleTime
	} else if len(releases) > 0 {
		releases[len(releases)-1].CycleTime = 0
		cycleTimeCount -= 1
	}
	averageCycleTime := float64(0)
	if cycleTimeCount > 0 {
		averageCycleTime = totalCycleTime / float64(cycleTimeCount)
	}

	metrics := &dto.Metrics{
		Series: releases,
		//ChangeFailureRate: changeFailureRate,
		AverageCycleTime: averageCycleTime,
	}

	if leadTimesCount > 0 {
		metrics.AverageLeadTime = totalLeadTime / float64(leadTimesCount)
	}

	impl.calculateChangeFailureRateAndRecoveryTime(metrics)
	if len(metrics.Series) > 0 {
		impl.calculateChangeSize(metrics)
	}
	return metrics, nil
}

func (impl DeploymentMetricServiceImpl) calculateChangeFailureRateAndRecoveryTime(metrics *dto.Metrics) {
	releases := metrics.Series
	failed := 0
	success := 0
	recoveryTime := float64(0)
	recovered := 0
	for _, v := range releases {
		if v.ReleaseStatus == dto.Failure {
			if metrics.LastFailedTime == "" {
				metrics.LastFailedTime = v.ReleaseTime.Format(layout)
			}
			//if i != 0 {
			//	releases[i].RecoveryTime = releases[i].ReleaseTime.Sub(releases[i+1].ReleaseTime)
			//	recoveryTime += int(releases[i].RecoveryTime.Hours())
			//}
			failed++
		}
		if v.ReleaseStatus == dto.Success {
			success++
		}
	}
	for i := 0; i < len(releases); i++ {
		if releases[i].ReleaseStatus == dto.Failure {
			if i < len(releases)-1 && releases[i+1].ReleaseStatus == dto.Failure {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				if releases[j].ReleaseStatus == dto.Success {
					releases[i].RecoveryTime = releases[j].ReleaseTime.Sub(releases[i].ReleaseTime).Minutes()
					recoveryTime += releases[i].RecoveryTime
					recovered++
					if metrics.RecoveryTimeLastFailed == 0 {
						metrics.RecoveryTimeLastFailed = releases[i].RecoveryTime
					}
					break
				}
			}
		}
	}
	changeFailureRate := float64(0)
	averageRecoveryTime := float64(0)
	if success+failed > 0 {
		changeFailureRate = float64(failed) * float64(100) / float64(failed+success)
	}
	if failed > 0 && recovered > 0 {
		averageRecoveryTime = recoveryTime / float64(recovered)
	}
	metrics.ChangeFailureRate = changeFailureRate
	metrics.AverageRecoveryTime = averageRecoveryTime
}

func (impl DeploymentMetricServiceImpl) calculateChangeSize(metrics *dto.Metrics) {
	releases := metrics.Series
	lineAdded := 0
	lineDeleted := 0
	deploymentSize := 0
	for _, v := range releases {
		lineAdded += v.ChangeSizeLineAdded
		lineDeleted += v.ChangeSizeLineDeleted
		deploymentSize += v.DeploymentSize
	}
	metrics.AverageDeploymentSize = float32(deploymentSize) / float32(len(releases))
	metrics.AverageLineAdded = float32(lineAdded) / float32(len(releases))
	metrics.AverageLineDeleted = float32(lineDeleted) / float32(len(releases))
}

func (impl DeploymentMetricServiceImpl) transform(releases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime) []*dto.Metric {
	pm := make(map[int]*sql.PipelineMaterial)
	for _, v := range materials {
		pm[v.AppReleaseId] = v
	}
	lt := make(map[int]sql.LeadTime)
	for _, v := range leadTimes {
		lt[v.AppReleaseId] = v
	}

	impl.logger.Errorw("materials ", "mat", pm)

	metrics := make([]*dto.Metric, 0)
	for _, v := range releases {
		metric := &dto.Metric{
			ReleaseType:           v.ReleaseType,
			ReleaseStatus:         v.ReleaseStatus,
			ReleaseTime:           v.TriggerTime,
			ChangeSizeLineAdded:   v.ChangeSizeLineAdded,
			ChangeSizeLineDeleted: v.ChangeSizeLineDeleted,
			DeploymentSize:        v.ChangeSizeLineDeleted + v.ChangeSizeLineAdded,
			LeadTime:              0,
			CycleTime:             0,
			RecoveryTime:          0,
		}
		if p, ok := pm[v.Id]; ok {
			metric.CommitHash = p.CommitHash
		} else {
			impl.logger.Errorf("not found appId: %d", v.AppId)
		}
		if l, ok := lt[v.Id]; ok {
			metric.LeadTime = l.LeadTime.Minutes()
		} else {
			impl.logger.Errorf("not found appId: %d", v.AppId)
		}
		metrics = append(metrics, metric)
	}
	return metrics
}
