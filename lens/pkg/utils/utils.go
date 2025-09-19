package utils

import (
	"fmt"
	"time"

	"github.com/devtron-labs/lens/internal/dto"
	"github.com/devtron-labs/lens/pkg/constants"
)

// ParseDateRange parses from and to date strings
func ParseDateRange(from, to string) (time.Time, time.Time, error) {
	fromTime, err := time.Parse(constants.Layout, from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	toTime, err := time.Parse(constants.Layout, to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return fromTime, toTime, nil
}

// GenerateAppEnvKey creates a consistent key for app-env pair mapping
func GenerateAppEnvKey(appId, envId int) string {
	return fmt.Sprintf("%d-%d", appId, envId)
}

// CreateEmptyMetrics creates an empty metrics response
func CreateEmptyMetrics() *dto.Metrics {
	return &dto.Metrics{Series: []*dto.Metric{}}
}
