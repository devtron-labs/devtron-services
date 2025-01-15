package adapter

import (
	"github.com/devtron-labs/ci-runner/executor/stage/bean"
	"strings"
)

func GetTargetPlatformObjectForList(targetPlatform string) []bean.TargetPlatform {
	targetPlatforms := GetTargetPlatformListFromString(targetPlatform)
	var targetPlatformObject []bean.TargetPlatform
	for _, targetPlatform := range targetPlatforms {
		targetPlatformObject = append(targetPlatformObject, bean.TargetPlatform{Name: targetPlatform})
	}
	return targetPlatformObject
}

func GetTargetPlatformListFromString(targetPlatform string) []string {
	return strings.Split(targetPlatform, ",")
}
