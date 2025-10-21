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
	"time"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/lens/internal/dto"
	"github.com/devtron-labs/lens/internal/sql"
	"github.com/devtron-labs/lens/pkg/constants"
	utils2 "github.com/devtron-labs/lens/pkg/utils"
	"go.uber.org/zap"
)

type DeploymentMetricService interface {
	GetDeploymentMetrics(request *dto.MetricRequest) (*dto.Metrics, error)
	GetBulkDeploymentMetrics(request *dto.BulkMetricRequest) (*dto.BulkMetricsResponse, error)

	// New DORA metrics functions
	ProcessSingleDoraMetrics(request *dto.MetricRequest) (*dto.Metrics, error)
	ProcessBulkDoraMetrics(request *dto.BulkMetricRequest) ([]DoraMetrics, error)
	CalculateDoraMetrics(appId, envId int, releases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, fromTime, toTime time.Time) *DoraMetrics
	GetDoraMetricsSummary(doraMetrics *DoraMetrics) *DoraMetricsSummary
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
	from, to, err := utils2.ParseDateRange(request.From, request.To)
	if err != nil {
		return nil, err
	}

	releases, err := impl.appReleaseRepository.GetReleaseBetween(request.AppId, request.EnvId, from, to)
	if err != nil {
		impl.logger.Errorw("error getting data from db ", "err", err)
		return nil, err
	}

	if len(releases) == 0 {
		return utils2.CreateEmptyMetrics(), nil
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

	return impl.getBulkDeploymentMetricsWithBulkQueries(request)
}

func (impl DeploymentMetricServiceImpl) getBulkDeploymentMetricsWithBulkQueries(request *dto.BulkMetricRequest) (*dto.BulkMetricsResponse, error) {
	response := &dto.BulkMetricsResponse{
		Results: make([]dto.AppEnvMetrics, len(request.AppEnvPairs)),
	}
	// Step 1: Get all releases for all app-env pairs in one query
	allReleases, err := impl.appReleaseRepository.GetReleaseBetweenBulk(request.AppEnvPairs, *request.From, *request.To)
	if err != nil {
		impl.logger.Errorw("error getting bulk releases from db", "err", err)
		return nil, err
	}

	// Step 2: Group releases by app-env pair
	releasesByAppEnv := make(map[string][]sql.AppRelease)
	var allReleaseIds []int

	for _, release := range allReleases {
		key := utils2.GenerateAppEnvKey(release.AppId, release.EnvironmentId)
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
		key := utils2.GenerateAppEnvKey(pair.AppId, pair.EnvId)
		releases := releasesByAppEnv[key]

		appEnvMetric := dto.AppEnvMetrics{
			AppId: pair.AppId,
			EnvId: pair.EnvId,
		}

		if len(releases) == 0 {
			appEnvMetric.Metrics = utils2.CreateEmptyMetrics()
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
				metrics.LastFailedTime = v.ReleaseTime.Format(constants.Layout)
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

// ============================================================================
// NEW DORA METRICS CALCULATION FUNCTIONS
// ============================================================================

// DoraMetrics represents the four key DORA metrics
type DoraMetrics struct {
	AppId                  int     `json:"app_id"`
	EnvId                  int     `json:"env_id"`
	DeploymentFrequency    float64 `json:"deployment_frequency"`       // Deployments per day
	ChangeFailureRate      float64 `json:"change_failure_rate"`        // Percentage
	MeanLeadTimeForChanges float64 `json:"mean_lead_time_for_changes"` // Minutes
	MeanTimeToRecovery     float64 `json:"mean_time_to_recovery"`      // Minutes
}

// CalculateDoraMetrics calculates all four DORA metrics based on the provided formulas
func (impl DeploymentMetricServiceImpl) CalculateDoraMetrics(appId, envId int, releases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, fromTime, toTime time.Time) *DoraMetrics {
	if len(releases) == 0 {
		return &DoraMetrics{
			AppId: appId,
			EnvId: envId,
		}
	}

	// Transform releases to dto.Metric format for easier processing
	metrics := impl.transform(releases, materials, leadTimes)

	return &DoraMetrics{
		AppId:                  appId,
		EnvId:                  envId,
		DeploymentFrequency:    impl.calculateDeploymentFrequency(metrics, fromTime, toTime),
		ChangeFailureRate:      impl.calculateChangeFailureRateNew(metrics),
		MeanLeadTimeForChanges: impl.calculateMeanLeadTimeForChanges(metrics),
		MeanTimeToRecovery:     impl.calculateMeanTimeToRecovery(metrics),
	}
}

// calculateDeploymentFrequency calculates deployment frequency
// Formula: Deployments to Production ÷ Time Period
func (impl DeploymentMetricServiceImpl) calculateDeploymentFrequency(metrics []*dto.Metric, fromTime, toTime time.Time) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	// calculating time period in days
	timePeriodDays := toTime.Sub(fromTime).Hours() / 24.0
	if timePeriodDays <= 0 {
		return 0.0
	}

	return float64(len(metrics)) / timePeriodDays
}

// calculateChangeFailureRateNew calculates change failure rate
// Formula: (Failed Deployments ÷ Total Deployments) × 100
func (impl DeploymentMetricServiceImpl) calculateChangeFailureRateNew(metrics []*dto.Metric) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	failedDeployments := 0
	totalDeployments := len(metrics)

	for _, metric := range metrics {
		if metric.ReleaseStatus == dto.Failure {
			failedDeployments++
		}
	}

	if totalDeployments == 0 {
		return 0.0
	}

	return (float64(failedDeployments) / float64(totalDeployments)) * 100.0
}

// calculateMeanLeadTimeForChanges calculates mean lead time for changes
// Formula: (Σ (Deployment Time – Commit Time)) ÷ Number of Changes
func (impl DeploymentMetricServiceImpl) calculateMeanLeadTimeForChanges(metrics []*dto.Metric) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	totalLeadTime := 0.0
	validChanges := 0

	for _, metric := range metrics {
		// Only consider successful deployments with valid lead time
		if metric.ReleaseStatus == dto.Success && metric.LeadTime != float64(0) {
			totalLeadTime += metric.LeadTime
			validChanges++
		}
	}

	if validChanges == 0 {
		return 0.0
	}

	return totalLeadTime / float64(validChanges)
}

// calculateMeanTimeToRecovery calculates mean time to recovery (MTTR)
// Formula: (Σ (Recovery Time – Failure Time)) ÷ Number of Incidents
func (impl DeploymentMetricServiceImpl) calculateMeanTimeToRecovery(metrics []*dto.Metric) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	totalRecoveryTime := 0.0
	incidents := 0

	// Sort metrics by release time (assuming they are already sorted by ID desc)
	// We need to find failed deployments and their recovery times
	for i := 0; i < len(metrics); i++ {
		if metrics[i].ReleaseStatus == dto.Failure {
			// Look for the next successful deployment after this failure
			recoveryTime := impl.findRecoveryTime(metrics, i)
			if recoveryTime > 0 {
				totalRecoveryTime += recoveryTime
				incidents++
			}
		}
	}

	if incidents == 0 {
		return 0.0
	}

	return totalRecoveryTime / float64(incidents)
}

// findRecoveryTime finds the recovery time for a failed deployment
func (impl DeploymentMetricServiceImpl) findRecoveryTime(metrics []*dto.Metric, failureIndex int) float64 {
	if failureIndex >= len(metrics) {
		return 0.0
	}

	failureTime := metrics[failureIndex].ReleaseTime

	// Look for the next successful deployment (going backwards in time since metrics are sorted by ID desc)
	for i := failureIndex - 1; i >= 0; i-- {
		if metrics[i].ReleaseStatus == dto.Success {
			recoveryTime := metrics[i].ReleaseTime
			// Calculate recovery time in minutes
			return recoveryTime.Sub(failureTime).Minutes()
		}
	}

	return 0.0
}

// CalculateDoraMetricsForBulk calculates DORA metrics for bulk processing (without materials data)
func (impl DeploymentMetricServiceImpl) CalculateDoraMetricsForBulk(appEnvPairs []dto.AppEnvPair, releasesByAppEnv map[string][]sql.AppRelease, allLeadTimes []sql.LeadTime, fromTime, toTime time.Time) []DoraMetrics {
	result := make([]DoraMetrics, len(appEnvPairs))

	for i, pair := range appEnvPairs {
		key := utils2.GenerateAppEnvKey(pair.AppId, pair.EnvId)
		releases := releasesByAppEnv[key]

		if len(releases) == 0 {
			result[i] = DoraMetrics{
				AppId: pair.AppId,
				EnvId: pair.EnvId,
			}
			continue
		}

		// Get release IDs for filtering lead times
		releaseIds := make([]int, len(releases))
		for j, release := range releases {
			releaseIds[j] = release.Id
		}

		// Filter lead times for this app-env pair
		leadTimes := impl.filterLeadTimesByReleaseIds(allLeadTimes, releaseIds)

		// Calculate DORA metrics for this app-env pair (without materials)
		doraMetrics := impl.CalculateDoraMetricsForBulkWithoutMaterials(pair.AppId, pair.EnvId, releases, leadTimes, fromTime, toTime)
		result[i] = *doraMetrics
	}

	return result
}

// CalculateDoraMetricsForBulkWithoutMaterials calculates DORA metrics without materials data (for bulk processing)
func (impl DeploymentMetricServiceImpl) CalculateDoraMetricsForBulkWithoutMaterials(appId, envId int, releases []sql.AppRelease, leadTimes []sql.LeadTime, fromTime, toTime time.Time) *DoraMetrics {
	if len(releases) == 0 {
		return &DoraMetrics{
			AppId: appId,
			EnvId: envId,
		}
	}

	// Transform releases to dto.Metric format without materials
	metrics := impl.transformWithoutMaterials(releases, leadTimes)

	return &DoraMetrics{
		AppId:                  appId,
		EnvId:                  envId,
		DeploymentFrequency:    impl.calculateDeploymentFrequency(metrics, fromTime, toTime),
		ChangeFailureRate:      impl.calculateChangeFailureRateNew(metrics),
		MeanLeadTimeForChanges: impl.calculateMeanLeadTimeForChanges(metrics),
		MeanTimeToRecovery:     impl.calculateMeanTimeToRecovery(metrics),
	}
}

// transformWithoutMaterials transforms releases to metrics without materials data
func (impl DeploymentMetricServiceImpl) transformWithoutMaterials(releases []sql.AppRelease, leadTimes []sql.LeadTime) []*dto.Metric {
	lt := make(map[int]sql.LeadTime)
	for _, v := range leadTimes {
		lt[v.AppReleaseId] = v
	}

	metrics := make([]*dto.Metric, 0)
	for _, v := range releases {
		metric := &dto.Metric{
			ReleaseType:   v.ReleaseType,
			ReleaseStatus: v.ReleaseStatus,
			ReleaseTime:   v.TriggerTime,
		}

		if l, ok := lt[v.Id]; ok {
			metric.LeadTime = l.LeadTime.Minutes()
		}

		metrics = append(metrics, metric)
	}
	return metrics
}

// CalculateDoraMetricsWithProductionFilter calculates DORA metrics with production environment filtering
func (impl DeploymentMetricServiceImpl) CalculateDoraMetricsWithProductionFilter(appId, envId int, releases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, fromTime, toTime time.Time, isProductionEnv bool) *DoraMetrics {
	if len(releases) == 0 {
		return &DoraMetrics{
			AppId: appId,
			EnvId: envId,
		}
	}

	// Transform releases to dto.Metric format
	metrics := impl.transform(releases, materials, leadTimes)

	return &DoraMetrics{
		AppId:                  appId,
		EnvId:                  envId,
		DeploymentFrequency:    impl.calculateDeploymentFrequencyWithFilter(metrics, fromTime, toTime, isProductionEnv),
		ChangeFailureRate:      impl.calculateChangeFailureRateNew(metrics),
		MeanLeadTimeForChanges: impl.calculateMeanLeadTimeForChanges(metrics),
		MeanTimeToRecovery:     impl.calculateMeanTimeToRecovery(metrics),
	}
}

// calculateDeploymentFrequencyWithFilter calculates deployment frequency with production filter
func (impl DeploymentMetricServiceImpl) calculateDeploymentFrequencyWithFilter(metrics []*dto.Metric, fromTime, toTime time.Time, isProductionEnv bool) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	// Count deployments based on environment type
	deploymentCount := 0
	for _, metric := range metrics {
		// For production environments, count all deployments
		// For non-production, count successful deployments only
		if isProductionEnv {
			deploymentCount++
		} else if metric.ReleaseStatus == dto.Success {
			deploymentCount++
		}
	}

	// Calculate time period in days
	timePeriodDays := toTime.Sub(fromTime).Hours() / 24.0
	if timePeriodDays <= 0 {
		return 0.0
	}

	return float64(deploymentCount) / timePeriodDays
}

// GetDoraMetricsSummary provides a summary of DORA metrics with performance classification
func (impl DeploymentMetricServiceImpl) GetDoraMetricsSummary(doraMetrics *DoraMetrics) *DoraMetricsSummary {
	return &DoraMetricsSummary{
		Metrics: doraMetrics,
		Performance: DoraPerformanceClassification{
			DeploymentFrequencyLevel: impl.classifyDeploymentFrequency(doraMetrics.DeploymentFrequency),
			ChangeFailureRateLevel:   impl.classifyChangeFailureRate(doraMetrics.ChangeFailureRate),
			MeanLeadTimeLevel:        impl.classifyMeanLeadTime(doraMetrics.MeanLeadTimeForChanges),
			MeanTimeToRecoveryLevel:  impl.classifyMeanTimeToRecovery(doraMetrics.MeanTimeToRecovery),
		},
	}
}

// DoraMetricsSummary contains DORA metrics with performance classification
type DoraMetricsSummary struct {
	Metrics     *DoraMetrics                  `json:"metrics"`
	Performance DoraPerformanceClassification `json:"performance"`
}

// DoraPerformanceClassification classifies DORA metrics performance levels
type DoraPerformanceClassification struct {
	DeploymentFrequencyLevel string `json:"deployment_frequency_level"`
	ChangeFailureRateLevel   string `json:"change_failure_rate_level"`
	MeanLeadTimeLevel        string `json:"mean_lead_time_level"`
	MeanTimeToRecoveryLevel  string `json:"mean_time_to_recovery_level"`
}

// Performance level constants
const (
	PerformanceElite  = "Elite"
	PerformanceHigh   = "High"
	PerformanceMedium = "Medium"
	PerformanceLow    = "Low"
)

// classifyDeploymentFrequency classifies deployment frequency performance
func (impl DeploymentMetricServiceImpl) classifyDeploymentFrequency(frequency float64) string {
	// Based on DORA research benchmarks (deployments per day)
	if frequency >= 1.0 {
		return PerformanceElite // Multiple deployments per day
	} else if frequency >= 0.14 { // ~1 per week
		return PerformanceHigh
	} else if frequency >= 0.033 { // ~1 per month
		return PerformanceMedium
	}
	return PerformanceLow
}

// classifyChangeFailureRate classifies change failure rate performance
func (impl DeploymentMetricServiceImpl) classifyChangeFailureRate(rate float64) string {
	// Based on DORA research benchmarks (percentage)
	if rate <= 15.0 {
		return PerformanceElite
	} else if rate <= 20.0 {
		return PerformanceHigh
	} else if rate <= 30.0 {
		return PerformanceMedium
	}
	return PerformanceLow
}

// classifyMeanLeadTime classifies mean lead time performance
func (impl DeploymentMetricServiceImpl) classifyMeanLeadTime(leadTime float64) string {
	// Convert minutes to hours for classification
	leadTimeHours := leadTime / 60.0

	// Based on DORA research benchmarks
	if leadTimeHours <= 24.0 { // Less than one day
		return PerformanceElite
	} else if leadTimeHours <= 168.0 { // Less than one week
		return PerformanceHigh
	} else if leadTimeHours <= 720.0 { // Less than one month
		return PerformanceMedium
	}
	return PerformanceLow
}

// classifyMeanTimeToRecovery classifies MTTR performance
func (impl DeploymentMetricServiceImpl) classifyMeanTimeToRecovery(mttr float64) string {
	// Convert minutes to hours for classification
	mttrHours := mttr / 60.0

	// Based on DORA research benchmarks
	if mttrHours <= 1.0 { // Less than one hour
		return PerformanceElite
	} else if mttrHours <= 24.0 { // Less than one day
		return PerformanceHigh
	} else if mttrHours <= 168.0 { // Less than one week
		return PerformanceMedium
	}
	return PerformanceLow
}

// ProcessBulkDoraMetrics processes DORA metrics for bulk requests (without materials data)
// This function can be used alongside getBulkDeploymentMetricsWithBulkQueries
func (impl DeploymentMetricServiceImpl) ProcessBulkDoraMetrics(request *dto.BulkMetricRequest) ([]DoraMetrics, error) {
	if len(request.AppEnvPairs) == 0 {
		return []DoraMetrics{}, nil
	}

	// Step 1: Get all releases for all app-env pairs in one query
	allReleases, err := impl.appReleaseRepository.GetReleaseBetweenBulk(request.AppEnvPairs, *request.From, *request.To)
	if err != nil {
		impl.logger.Errorw("error getting bulk releases for DORA metrics", "err", err)
		return nil, err
	}

	// Step 2: Group releases by app-env pair
	releasesByAppEnv := make(map[string][]sql.AppRelease)
	var allReleaseIds []int

	for _, release := range allReleases {
		key := utils2.GenerateAppEnvKey(release.AppId, release.EnvironmentId)
		releasesByAppEnv[key] = append(releasesByAppEnv[key], release)
		allReleaseIds = append(allReleaseIds, release.Id)
	}

	// Step 3: Get only lead times in bulk (materials not needed for bulk processing)
	var allLeadTimes []sql.LeadTime

	if len(allReleaseIds) > 0 {
		allLeadTimes, err = impl.leadTimeRepository.FindByIds(allReleaseIds)
		if err != nil {
			impl.logger.Errorw("error getting bulk lead times for DORA metrics", "err", err)
			return nil, err
		}
	}

	// Step 4: Calculate DORA metrics for all app-env pairs (without materials)
	return impl.CalculateDoraMetricsForBulk(request.AppEnvPairs, releasesByAppEnv, allLeadTimes, *request.From, *request.To), nil
}

// calculateCycleTimeBetweenReleases calculates the time between consecutive releases
func (impl DeploymentMetricServiceImpl) calculateCycleTimeBetweenReleases(releases []*dto.Metric, lastRelease *sql.AppRelease) {
	if len(releases) == 0 {
		return
	}

	// Calculate cycle time between consecutive releases
	for i := 0; i < len(releases)-1; i++ {
		releases[i].CycleTime = releases[i].ReleaseTime.Sub(releases[i+1].ReleaseTime).Minutes()
	}

	// Handle the last release
	if lastRelease != nil {
		releases[len(releases)-1].CycleTime = releases[len(releases)-1].ReleaseTime.Sub(lastRelease.TriggerTime).Minutes()
	} else if len(releases) > 0 {
		releases[len(releases)-1].CycleTime = 0
	}
}

// populateMetricsWithImprovedLogic populates dto.Metrics using DORA calculation helper functions
func (impl DeploymentMetricServiceImpl) populateMetricsWithImprovedLogic(appReleases []sql.AppRelease, materials []*sql.PipelineMaterial, leadTimes []sql.LeadTime, lastRelease *sql.AppRelease, fromTime, toTime time.Time) (*dto.Metrics, error) {
	releases := impl.transform(appReleases, materials, leadTimes)

	impl.calculateCycleTimeBetweenReleases(releases, lastRelease)

	deploymentFrequency := impl.calculateDeploymentFrequency(releases, fromTime, toTime)
	averageLeadTime := impl.calculateMeanLeadTimeForChanges(releases)
	changeFailureRate := impl.calculateChangeFailureRateNew(releases)
	averageRecoveryTime := impl.calculateMeanTimeToRecovery(releases)

	lastFailedTime := ""
	recoveryTimeLastFailed := float64(0)
	for i := 0; i < len(releases); i++ {
		if releases[i].ReleaseStatus == dto.Failure {
			if lastFailedTime == "" {
				lastFailedTime = releases[i].ReleaseTime.Format(constants.Layout)
			}
			if i < len(releases)-1 && releases[i+1].ReleaseStatus == dto.Failure {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				if releases[j].ReleaseStatus == dto.Success {
					releases[i].RecoveryTime = releases[j].ReleaseTime.Sub(releases[i].ReleaseTime).Minutes()
					if recoveryTimeLastFailed == 0 {
						recoveryTimeLastFailed = releases[i].RecoveryTime
					}
					break
				}
			}
		}
	}

	metrics := &dto.Metrics{
		Series:                 releases,
		AverageCycleTime:       deploymentFrequency,
		AverageLeadTime:        averageLeadTime,
		ChangeFailureRate:      changeFailureRate,
		AverageRecoveryTime:    averageRecoveryTime,
		LastFailedTime:         lastFailedTime,
		RecoveryTimeLastFailed: recoveryTimeLastFailed,
	}

	// Calculate change size metrics
	if len(metrics.Series) > 0 {
		impl.calculateChangeSize(metrics)
	}

	return metrics, nil
}

// ProcessSingleDoraMetrics processes DORA metrics for a single app-env pair
func (impl DeploymentMetricServiceImpl) ProcessSingleDoraMetrics(request *dto.MetricRequest) (*dto.Metrics, error) {
	from, to, err := utils2.ParseDateRange(request.From, request.To)
	if err != nil {
		return nil, err
	}

	releases, err := impl.appReleaseRepository.GetReleaseBetween(request.AppId, request.EnvId, from, to)
	if err != nil {
		impl.logger.Errorw("error getting releases for DORA metrics", "err", err)
		return nil, err
	}

	if len(releases) == 0 {
		return utils2.CreateEmptyMetrics(), nil
	}

	var releaseIds []int
	for _, v := range releases {
		releaseIds = append(releaseIds, v.Id)
	}

	materials, err := impl.pipelineMaterialRepository.FindByAppReleaseIds(releaseIds)
	if err != nil {
		impl.logger.Errorw("error getting materials for DORA metrics", "err", err)
		return nil, err
	}

	leadTimes, err := impl.leadTimeRepository.FindByIds(releaseIds)
	if err != nil {
		impl.logger.Errorw("error getting lead times for DORA metrics", "err", err)
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

	return impl.populateMetricsWithImprovedLogic(releases, materials, leadTimes, lastRelease, from, to)
}
