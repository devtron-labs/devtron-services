package adaptor

import (
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/common-lib/constants"
)

func GetImageScanEvent(dest, digest string, commonWorkflowRequest *helper.CommonWorkflowRequest) *helper.ScanEvent {
	if commonWorkflowRequest == nil {
		return &helper.ScanEvent{}
	}
	return &helper.ScanEvent{
		Image:               dest,
		ImageDigest:         digest,
		PipelineId:          commonWorkflowRequest.PipelineId,
		UserId:              commonWorkflowRequest.TriggeredBy,
		DockerRegistryId:    commonWorkflowRequest.DockerRegistryId,
		DockerConnection:    commonWorkflowRequest.DockerConnection,
		DockerCert:          commonWorkflowRequest.DockerCert,
		ImageScanMaxRetries: commonWorkflowRequest.ImageScanMaxRetries,
		ImageScanRetryDelay: commonWorkflowRequest.ImageScanRetryDelay,
		SourceType:          constants.SourceTypeImage,
		SourceSubType:       constants.SourceSubTypeCi,
	}
}
