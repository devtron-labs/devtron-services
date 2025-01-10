package security

import (
	"github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/image-scanner/pkg/sql/repository"
)

type ScanCodeRequest struct {
	ScanEvent               *bean.ImageScanEvent
	Tool                    *repository.ScanToolMetadata
	ExecutionHistory        *repository.ImageScanExecutionHistory
	ExecutionHistoryDirPath string
}
