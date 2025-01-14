package adaptor

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	bean2 "github.com/devtron-labs/image-scanner/pkg/sql/bean"
	"github.com/devtron-labs/image-scanner/pkg/sql/repository"
	"time"
)

func GetImageScanExecutionHistory(imageScanEvent *bean.ImageScanEvent, scanEventJson []byte, time time.Time) *repository.ImageScanExecutionHistory {
	return &repository.ImageScanExecutionHistory{
		Image:              imageScanEvent.Image,
		ImageHash:          imageScanEvent.ImageDigest,
		ExecutionTime:      time,
		ExecutedBy:         imageScanEvent.UserId,
		SourceMetadataJson: string(scanEventJson),
		SourceType:         imageScanEvent.SourceType,
		SourceSubType:      imageScanEvent.SourceSubType,
	}
}

func GetScanToolExecutionHistoryMapping(executionHistoryModel *repository.ImageScanExecutionHistory, state bean2.ScanExecutionProcessState,
	errorMsg string, toolId int) *repository.ScanToolExecutionHistoryMapping {
	return &repository.ScanToolExecutionHistoryMapping{
		ImageScanExecutionHistoryId: executionHistoryModel.Id,
		ScanToolId:                  toolId,
		ExecutionStartTime:          executionHistoryModel.ExecutionTime,
		State:                       state,
		ErrorMessage:                errorMsg,
		AuditLog: repository.AuditLog{
			CreatedOn: executionHistoryModel.ExecutionTime,
			CreatedBy: int32(executionHistoryModel.ExecutedBy),
			UpdatedOn: executionHistoryModel.ExecutionTime,
			UpdatedBy: int32(executionHistoryModel.ExecutedBy),
		},
	}
}

func GetResourceScanExecutionResult(executionHistoryId, scanToolId int, scanDataJson string, format repository.ResourceScanFormat, types []int) *repository.ResourceScanResult {
	return &repository.ResourceScanResult{
		ImageScanExecutionHistoryId: executionHistoryId,
		ScanDataJson:                scanDataJson,
		Format:                      format,
		Types:                       types,
		ScanToolId:                  scanToolId,
	}
}
