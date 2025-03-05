/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package constants

import "fmt"

/**
 	Cluster - 			1000-1999
	Environment - 		2000-2999
	Global Config - 	3000-3999
	Pipeline Config -	4000-4999
	Pipeline - 			5000-5999
	User - 				6000-6999
	Other -				7000-7999
*/

type ErrorCode struct {
	Code           string
	userErrMessage string
}

func (code ErrorCode) UserMessage(params ...interface{}) string {
	return fmt.Sprintf(code.userErrMessage, params)
}

const (
	//Cluster Errors
	ClusterCreateDBFailed      string = "1001"
	ClusterCreateACDFailed     string = "1002"
	ClusterDBRollbackFailed    string = "1003"
	ClusterUpdateDBFailed      string = "1004"
	ClusterUpdateACDFailed     string = "1005"
	ClusterCreateBadRequestACD string = "1006"
	ClusterUpdateBadRequestACD string = "1007"

	//Environment Errors
	EnvironmentCreateDBFailed          string = "2001"
	EnvironmentUpdateDBFailed          string = "2002"
	EnvironmentUpdateEnvOverrideFailed string = "2003"

	//Global Config Errors Constants; use 3000
	DockerRegCreateFailedInDb   string = "3001"
	DockerRegCreateFailedInGocd string = "3002"
	DockerRegUpdateFailedInDb   string = "3003"
	DockerRegUpdateFailedInGocd string = "3004"

	GitProviderCreateFailedAlreadyExists string = "3005"
	GitProviderCreateFailedInDb          string = "3006"
	GitProviderUpdateProviderNotExists   string = "3007"
	GitProviderUpdateFailedInDb          string = "3008"
	DockerRegDeleteFailedInDb            string = "3009"
	DockerRegDeleteFailedInGocd          string = "3010"
	GitProviderUpdateFailedInSync        string = "3011"
	GitProviderUpdateRequestIsInvalid    string = "3012"
	// For conflicts use 900 series
	GitOpsConfigValidationConflict string = "3900"

	ChartCreatedAlreadyExists string = "5001"
	ChartNameAlreadyReserved  string = "5002"

	UserCreateDBFailed        string = "6001"
	UserCreatePolicyFailed    string = "6002"
	UserUpdateDBFailed        string = "6003"
	UserUpdatePolicyFailed    string = "6004"
	UserNoTokenProvided       string = "6005"
	UserNotFoundForToken      string = "6006"
	UserCreateFetchRoleFailed string = "6007"
	UserUpdateFetchRoleFailed string = "6008"

	AppDetailResourceTreeNotFound string = "7000"
	HelmReleaseNotFound           string = "7001"

	CasbinPolicyNotCreated string = "8000"

	GitHostCreateFailedAlreadyExists string = "9001"
	GitHostCreateFailedInDb          string = "9002"

	// feasibility errors
	VulnerabilityFound                   string = "10001"
	ApprovalNodeFail                     string = "10002"
	FilteringConditionFail               string = "10003"
	DeploymentWindowFail                 string = "10004"
	PreCDDoesNotExists                   string = "10005"
	PostCDDoesNotExists                  string = "10006"
	ArtifactNotAvailable                 string = "10007"
	DeploymentWindowByPassed             string = "10008"
	MandatoryPluginNotAdded              string = "10009"
	MandatoryTagNotAdded                 string = "10010"
	SecurityScanFail                     string = "10011"
	ApprovalConfigDependentActionFailure string = "10012"
	//Not Processed Internal error
	NotProcessed string = "11001"
	NotExecuted  string = "11002"
)
const (
	HttpStatusUnprocessableEntity = "422"
)

const (
	HttpClientSideTimeout = 499
)

var AppAlreadyExists = &ErrorCode{"4001", "application %s already exists"}
var AppDoesNotExist = &ErrorCode{"4004", "application %s does not exist"}

const (
	ErrorDeletingPipelineForDeletedArgoAppMsg = "error in deleting devtron pipeline for deleted argocd app"
	ArgoAppDeletedErrMsg                      = "argocd app deleted"
	UnableToFetchResourceTreeErrMsg           = "unable to fetch resource tree"
	UnableToFetchResourceTreeForAcdErrMsg     = "app detail fetched, failed to get resource tree from acd"
	CannotGetAppWithRefreshErrMsg             = "cannot get application with refresh"
	NoDataFoundErrMsg                         = "no data found"
)
