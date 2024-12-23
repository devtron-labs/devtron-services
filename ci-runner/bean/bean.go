package bean

type ResourceScanSourceType int

const (
	SourceTypeImage ResourceScanSourceType = 1
)

type ResourceScanSourceSubType int

const (
	SourceSubTypeCi ResourceScanSourceSubType = 1 // ci built image(1,1)
)
