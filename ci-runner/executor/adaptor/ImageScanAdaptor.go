package adaptor

import (
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/imageScan/bean"
)

func GetImageScanEvent(dest, digest string, commonWorkflowRequest *helper.CommonWorkflowRequest) *helper.ScanEvent {
	if commonWorkflowRequest == nil {
		return &helper.ScanEvent{}
	}
	return &helper.ScanEvent{
		ImageScanEvent: bean.ImageScanEvent{
			Image:            dest,
			ImageDigest:      digest,
			PipelineId:       commonWorkflowRequest.PipelineId,
			UserId:           commonWorkflowRequest.TriggeredBy,
			DockerRegistryId: commonWorkflowRequest.DockerRegistryId,
			DockerConnection: commonWorkflowRequest.DockerConnection,
			DockerCert:       commonWorkflowRequest.DockerCert,
			SourceType:       constants.SourceTypeImage,
			SourceSubType:    constants.SourceSubTypeCi,
		},
		ImageScanMaxRetries: commonWorkflowRequest.ImageScanMaxRetries,
		ImageScanRetryDelay: commonWorkflowRequest.ImageScanRetryDelay,
	}
}
