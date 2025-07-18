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

package helper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/caarlos0/env"
	cicxt "github.com/devtron-labs/ci-runner/executor/context"
	bean2 "github.com/devtron-labs/ci-runner/helper/bean"
	"github.com/devtron-labs/ci-runner/util"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/devtron-labs/common-lib/utils/dockerOperations"
	"github.com/devtron-labs/common-lib/utils/retryFunc"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DEVTRON_ENV_VAR_PREFIX   = "$devtron_env_"
	BUILD_ARG_FLAG           = "--build-arg"
	ROOT_PATH                = "."
	BUILDX_K8S_DRIVER_NAME   = "devtron-buildx-builder"
	BUILDX_NODE_NAME         = "devtron-buildx-node-"
	DOCKERD_OUTPUT_FILE_PATH = "/usr/local/bin/nohup.out"
)

type DockerHelper interface {
	StartDockerDaemonAndDockerLogin(commonWorkflowRequest *CommonWorkflowRequest, isSubStep bool) error
	BuildArtifact(ciRequest *CommonWorkflowRequest) (string, error)
	StopDocker(ciContext cicxt.CiContext) error
	PushArtifact(ciContext cicxt.CiContext, dest string) error
	ExtractDigestForBuildx(dest string, ciRequest *CommonWorkflowRequest) (string, error)
	CleanBuildxK8sDriver(ciContext cicxt.CiContext, nodes []map[string]string) error
	GetDestForNatsEvent(commonWorkflowRequest *CommonWorkflowRequest, dest string) (string, error)
	ExtractDigestUsingPull(dest string) (string, error)
	ExtractDigestFromImage(image string, useDockerApiToGetDigest bool, dockerAuthConfig *bean.DockerAuthConfig) (string, error)
	GetDockerAuthConfigForPrivateRegistries(workflowRequest *CommonWorkflowRequest) *bean.DockerAuthConfig
}

type DockerHelperImpl struct {
	DockerCommandEnv []string
	cmdExecutor      CommandExecutor
}

func NewDockerHelperImpl(cmdExecutor CommandExecutor) *DockerHelperImpl {
	return &DockerHelperImpl{
		DockerCommandEnv: os.Environ(),
		cmdExecutor:      cmdExecutor,
	}
}

func (impl *DockerHelperImpl) GetDestForNatsEvent(commonWorkflowRequest *CommonWorkflowRequest, dest string) (string, error) {
	return dest, nil
}

func (impl *DockerHelperImpl) StartDockerDaemonAndDockerLogin(commonWorkflowRequest *CommonWorkflowRequest, isSubStep bool) error {
	startDockerDaemon := func() error {
		connection := commonWorkflowRequest.DockerConnection
		dockerRegistryUrl := commonWorkflowRequest.IntermediateDockerRegistryUrl
		registryUrl, err := util.ParseUrl(dockerRegistryUrl)
		if err != nil {
			return err
		}
		host := registryUrl.Host
		dockerdstart := ""
		defaultAddressPoolFlag := ""
		dockerMtuValueFlag := ""
		if len(commonWorkflowRequest.DefaultAddressPoolBaseCidr) > 0 {
			if commonWorkflowRequest.DefaultAddressPoolSize <= 0 {
				commonWorkflowRequest.DefaultAddressPoolSize = 24
			}
			defaultAddressPoolFlag = fmt.Sprintf("--default-address-pool base=%s,size=%d", commonWorkflowRequest.DefaultAddressPoolBaseCidr, commonWorkflowRequest.DefaultAddressPoolSize)
		}
		if commonWorkflowRequest.CiBuildDockerMtuValue > 0 {
			dockerMtuValueFlag = fmt.Sprintf("--mtu=%d", commonWorkflowRequest.CiBuildDockerMtuValue)
		}
		if connection == util.INSECURE {
			dockerdstart = fmt.Sprintf("dockerd  %s --insecure-registry %s --host=unix:///var/run/docker.sock %s --host=tcp://0.0.0.0:2375 > %s 2>&1 &", defaultAddressPoolFlag, host, dockerMtuValueFlag, DOCKERD_OUTPUT_FILE_PATH)
			log.Println("Insecure Registry")
		} else {
			if connection == util.SECUREWITHCERT {
				log.Println("Secure with Cert")

				// Create /etc/docker/certs.d/<host>/ca.crt with specified content
				certDir := fmt.Sprintf("%s/%s", CertDir, host)
				os.MkdirAll(certDir, os.ModePerm)
				certFilePath := fmt.Sprintf("%s/ca.crt", certDir)

				log.Printf("creating %s", certFilePath)

				if err := util.CreateAndWriteFile(certFilePath, commonWorkflowRequest.DockerCert); err != nil {
					return err
				}

				// Run "update-ca-certificates" to update the system certificates
				log.Println(UpdateCaCertCommand)
				cpCmd := exec.Command("cp", certFilePath, CaCertPath)
				if err := cpCmd.Run(); err != nil {
					return err
				}

				updateCmd := exec.Command(UpdateCaCertCommand)
				if err := updateCmd.Run(); err != nil {
					return err
				}

				// Create /etc/buildkitd.toml with specified content
				log.Printf("creating %s", BuildkitdConfigPath)
				buildkitdContent := util.GenerateBuildkitdContent(host)

				if err := util.CreateAndWriteFile(BuildkitdConfigPath, buildkitdContent); err != nil {
					return err
				}
			}
			dockerdstart = fmt.Sprintf("dockerd %s --host=unix:///var/run/docker.sock %s --host=tcp://0.0.0.0:2375 > %s 2>&1 &", defaultAddressPoolFlag, dockerMtuValueFlag, DOCKERD_OUTPUT_FILE_PATH)
		}
		cmd := impl.GetCommandToExecute(dockerdstart)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("failed to start docker daemon")
			util.PrintFileContent(DOCKERD_OUTPUT_FILE_PATH)
			return err
		}
		log.Println("docker daemon started ", string(out))
		err = impl.waitForDockerDaemon(util.DOCKER_PS_START_WAIT_SECONDS)
		if err != nil {
			util.PrintFileContent(DOCKERD_OUTPUT_FILE_PATH)
			return err
		}
		shouldDockerLogin := len(commonWorkflowRequest.IntermediateDockerRegistryUrl) != 0
		if shouldDockerLogin {
			ciContext := cicxt.BuildCiContext(context.Background(), commonWorkflowRequest.EnableSecretMasking)
			err = impl.DockerLogin(ciContext, &DockerCredentials{
				DockerUsername:     commonWorkflowRequest.DockerUsername,
				DockerPassword:     commonWorkflowRequest.DockerPassword,
				AwsRegion:          commonWorkflowRequest.AwsRegion,
				AccessKey:          commonWorkflowRequest.AccessKey,
				SecretKey:          commonWorkflowRequest.SecretKey,
				DockerRegistryURL:  commonWorkflowRequest.IntermediateDockerRegistryUrl,
				DockerRegistryType: commonWorkflowRequest.DockerRegistryType,
				CredentialsType:    commonWorkflowRequest.CredentialsType,
			})
		}
		if err != nil {
			return err
		}

		return nil
	}

	if isSubStep {
		return startDockerDaemon()
	} else {
		if err := util.ExecuteWithStageInfoLog(util.DOCKER_DAEMON, startDockerDaemon); err != nil {
			return err
		}
		return nil
	}

}

const CertDir = "/etc/docker/certs.d"
const UpdateCaCertCommand = "update-ca-certificates"
const CaCertPath = "/usr/local/share/ca-certificates/"
const BuildkitdConfigPath = "/etc/buildkitd.toml"
const DOCKER_REGISTRY_TYPE_ECR = "ecr"
const DOCKER_REGISTRY_TYPE_DOCKERHUB = "docker-hub"
const DOCKER_REGISTRY_TYPE_OTHER = "other"
const REGISTRY_TYPE_ARTIFACT_REGISTRY = "artifact-registry"
const REGISTRY_TYPE_GCR = "gcr"
const JSON_KEY_USERNAME = "_json_key"
const CacheModeMax = "max"
const CacheModeMin = "min"

type DockerCredentials struct {
	DockerUsername, DockerPassword, AwsRegion, AccessKey, SecretKey, DockerRegistryURL, DockerRegistryType, CredentialsType string
}

type EnvironmentVariables struct {
	ShowDockerBuildCmdInLogs bool `env:"SHOW_DOCKER_BUILD_ARGS" envDefault:"true"`
}

func (impl *DockerHelperImpl) GetCommandToExecute(cmd string) *exec.Cmd {
	execCmd := exec.Command("/bin/sh", "-c", cmd)
	execCmd.Env = append(execCmd.Env, impl.DockerCommandEnv...)
	return execCmd
}

func (impl *DockerHelperImpl) DockerLogin(ciContext cicxt.CiContext, dockerCredentials *DockerCredentials) error {
	if dockerCredentials.CredentialsType == string(constants.CredentialsTypeAnonymous) {
		return nil
	}
	performDockerLogin := func() error {
		username := dockerCredentials.DockerUsername
		pwd := dockerCredentials.DockerPassword
		if dockerCredentials.DockerRegistryType == DOCKER_REGISTRY_TYPE_ECR {
			accessKey, secretKey := dockerCredentials.AccessKey, dockerCredentials.SecretKey
			//fmt.Printf("accessKey %s, secretKey %s\n", accessKey, secretKey)

			var creds *credentials.Credentials

			if len(dockerCredentials.AccessKey) == 0 || len(dockerCredentials.SecretKey) == 0 {
				//fmt.Println("empty accessKey or secretKey")
				sess, err := session.NewSession(&aws.Config{
					Region: &dockerCredentials.AwsRegion,
				})
				if err != nil {
					log.Println(err)
					return err
				}
				creds = ec2rolecreds.NewCredentials(sess)
			} else {
				creds = credentials.NewStaticCredentials(accessKey, secretKey, "")
			}
			sess, err := session.NewSession(&aws.Config{
				Region:      &dockerCredentials.AwsRegion,
				Credentials: creds,
			})
			if err != nil {
				log.Println(err)
				return err
			}
			svc := ecr.New(sess)
			input := &ecr.GetAuthorizationTokenInput{}
			authData, err := svc.GetAuthorizationToken(input)
			if err != nil {
				log.Println(err)
				return err
			}
			// decode token
			token := authData.AuthorizationData[0].AuthorizationToken
			decodedToken, err := base64.StdEncoding.DecodeString(*token)
			if err != nil {
				log.Println(err)
				return err
			}
			credsSlice := strings.Split(string(decodedToken), ":")
			username = credsSlice[0]
			pwd = credsSlice[1]

		} else if (dockerCredentials.DockerRegistryType == REGISTRY_TYPE_GCR || dockerCredentials.DockerRegistryType == REGISTRY_TYPE_ARTIFACT_REGISTRY) && username == JSON_KEY_USERNAME {
			// for gcr and artifact registry password is already saved as string in DB
			if strings.HasPrefix(pwd, "'") {
				pwd = pwd[1:]
			}
			if strings.HasSuffix(pwd, "'") {
				pwd = pwd[:len(pwd)-1]
			}
		}
		host := dockerCredentials.DockerRegistryURL
		dockerLogin := fmt.Sprintf("docker login -u '%s' -p '%s' '%s' ", username, pwd, host)

		awsLoginCmd := impl.GetCommandToExecute(dockerLogin)
		err := impl.cmdExecutor.RunCommand(ciContext, awsLoginCmd)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Docker login successful with username ", username, " on docker registry URL ", dockerCredentials.DockerRegistryURL)
		return nil
	}

	return performDockerLogin()
}

func (impl *DockerHelperImpl) executeDockerReBuild(ciContext cicxt.CiContext, k8sClient BuildxK8sInterface,
	useBuildxK8sDriver bool, dockerBuild string, deploymentNames []string,
	dockerBuildStageMetadata bean2.DockerBuildStageMetadata, reBuildLogs []any) error {
	if !useBuildxK8sDriver {
		return nil
	}
	k8sErr := k8sClient.RestartBuilders(ciContext)
	if k8sErr != nil {
		log.Println(util.DEVTRON, fmt.Sprintf(" error in RestartBuilders : %s", k8sErr.Error()))
		return k8sErr
	}
	k8sClient, err := newBuildxK8sClient(deploymentNames)
	if err != nil {
		log.Println(util.DEVTRON, " error in creating buildxK8sClient , err : ", err.Error())
		return err
	}
	err = k8sClient.RegisterBuilderPods(ciContext)
	if err != nil {
		log.Println(util.DEVTRON, " error in registering builder pods ", " err: ", err.Error())
		return err
	}
	rebuildImageStage := func() error {
		// wait for the builder pod to be up again
		startTime := time.Now()
		util.LogInfo("Waiting for builder pod to be ready,", "timeout: 2 minutes")
		done := make(chan bool)
		ctx, cancel := context.WithCancel(ciContext)
		defer cancel()
		go k8sClient.WaitUntilBuilderPodLive(ctx, done)
		select {
		case <-done:
			// builder pod is up again, continue with the build
			cancel()
		case <-time.After(2 * time.Minute):
			// timeout after 2 minutes
			cancel()
			return BuilderPodDeletedError
		}
		util.LogInfo("DONE -->", time.Since(startTime).Seconds())
		buildImageFunc := impl.buildImageStage(ciContext, dockerBuild, useBuildxK8sDriver, k8sClient, reBuildLogs)
		if buildImageFunc != nil {
			return buildImageFunc()
		}
		return nil
	}
	err = util.ExecuteWithStageInfoLogWithMetadata(
		util.DOCKER_REBUILD,
		dockerBuildStageMetadata,
		rebuildImageStage,
	)
	if err != nil && !errors.Is(err, BuilderPodDeletedError) {
		return err
	} else if errors.Is(err, BuilderPodDeletedError) {
		// Log error message for builder pod interruption due to
		util.LogError(BuilderPodDeletedError)
		util.LogWarn("Frequent spot interruptions can lead to build failures.",
			"Consider using a different node type or increasing the spot interruption tolerance.")
		// if the builder pod is deleted, we will retry the build
		return retryFunc.NewRetryableError(BuilderPodDeletedError)
	}
	return nil
}

func (impl *DockerHelperImpl) buildImageStage(ciContext cicxt.CiContext, dockerBuild string,
	useBuildxK8sDriver bool, k8sClient BuildxK8sInterface, buildLogs []any) func() error {
	return func() error {
		ctx, cancel := context.WithCancel(ciContext)
		defer cancel()
		errGroup, groupCtx := errgroup.WithContext(ctx)
		errGroup.Go(func() error {
			if useBuildxK8sDriver && k8sClient != nil {
				if err := k8sClient.CatchBuilderPodLivenessError(groupCtx); err != nil {
					cancel() // Cancel the context if there's an error in catching builder pod liveness
					return err
				}
			}
			return nil
		})
		errGroup.Go(func() error {
			if len(buildLogs) > 0 {
				log.Println(buildLogs...)
			}
			err := impl.runDockerBuildCommand(cicxt.BuildCiContext(groupCtx, ciContext.EnableSecretMasking), dockerBuild)
			if err != nil {
				return err
			}
			cancel() // Cancel the context after the build is done
			return nil
		})
		if err := errGroup.Wait(); err != nil {
			return err
		}
		return nil
	}
}

func (impl *DockerHelperImpl) BuildArtifact(ciRequest *CommonWorkflowRequest) (string, error) {
	ciContext := cicxt.BuildCiContext(context.Background(), ciRequest.EnableSecretMasking)
	envVars := &EnvironmentVariables{}
	err := env.Parse(envVars)
	if err != nil {
		log.Println("Error while parsing environment variables", err)
	}
	if ciRequest.DockerImageTag == "" {
		ciRequest.DockerImageTag = "latest"
	}
	ciBuildConfig := ciRequest.CiBuildConfig
	// Docker build, tag image and push
	dockerFileLocationDir := ciRequest.CheckoutPath
	log.Println(util.DEVTRON, " docker file location: ", dockerFileLocationDir)

	dest, err := BuildDockerImagePath(ciRequest)
	if err != nil {
		return "", err
	}
	if ciBuildConfig.CiBuildType == SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfig.CiBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuild := "docker build "
		if ciRequest.CacheInvalidate && ciRequest.IsPvcMounted {
			dockerBuild = dockerBuild + "--no-cache "
		}
		dockerBuildConfig := ciBuildConfig.DockerBuildConfig

		useBuildx := dockerBuildConfig.CheckForBuildX()
		dockerBuildxBuild := "docker buildx build "
		if useBuildx {
			if ciRequest.CacheInvalidate && ciRequest.IsPvcMounted {
				dockerBuild = dockerBuildxBuild + "--no-cache "
			} else {
				dockerBuild = dockerBuildxBuild + " "
			}
		}
		dockerBuildFlags := getDockerBuildFlagsMap(dockerBuildConfig)
		for key, value := range dockerBuildFlags {
			dockerBuild = dockerBuild + " " + key + value
		}
		if !ciRequest.EnableBuildContext || dockerBuildConfig.BuildContext == "" {
			dockerBuildConfig.BuildContext = ROOT_PATH
		}
		dockerBuildConfig.BuildContext = path.Join(ROOT_PATH, dockerBuildConfig.BuildContext)

		dockerfilePath := getDockerfilePath(ciBuildConfig, ciRequest.CheckoutPath)
		var buildxExportCacheFunc func() error = nil
		useBuildxK8sDriver, eligibleK8sDriverNodes, deploymentNames := false, make([]map[string]string, 0), make([]string, 0)
		var k8sClient BuildxK8sInterface
		if useBuildx {
			setupBuildxBuilder := func() error {
				err := impl.checkAndCreateDirectory(ciContext, util.LOCAL_BUILDX_LOCATION)
				if err != nil {
					log.Println(util.DEVTRON, " error in creating LOCAL_BUILDX_LOCATION ", util.LOCAL_BUILDX_LOCATION)
					return err
				}
				useBuildxK8sDriver, eligibleK8sDriverNodes = dockerBuildConfig.CheckForBuildXK8sDriver()
				if useBuildxK8sDriver {
					deploymentNames, err = impl.createBuildxBuilderWithK8sDriver(ciContext, ciRequest.PropagateLabelsInBuildxPod, ciRequest.DockerConnection, dockerBuildConfig.BuildxDriverImage, eligibleK8sDriverNodes, ciRequest.PipelineId, ciRequest.WorkflowId)
					if err != nil {
						log.Println(util.DEVTRON, " error in creating buildxDriver , err : ", err.Error())
						return err
					}
					k8sClient, err = newBuildxK8sClient(deploymentNames)
					if err != nil {
						log.Println(util.DEVTRON, " error in creating buildxK8sClient , err : ", err.Error())
						return err
					}
					k8sClient.PatchOwnerReferenceInBuilders()
					err = k8sClient.RegisterBuilderPods(ciContext)
					if err != nil {
						log.Println(util.DEVTRON, " error in registering builder pods ", " err: ", err)
						return err
					}
				} else {
					err = impl.createBuildxBuilderForMultiArchBuild(ciContext, ciRequest.DockerConnection, dockerBuildConfig.BuildxDriverImage)
					if err != nil {
						return err
					}
				}
				return nil
			}

			if err = util.ExecuteWithStageInfoLog(util.SETUP_BUILDX_BUILDER, setupBuildxBuilder); err != nil {
				return "", err
			}

			cacheEnabled := (ciRequest.IsPvcMounted || ciRequest.BlobStorageConfigured)
			oldCacheBuildxPath, localCachePath := "", ""

			if cacheEnabled {
				log.Println(" -----> Setting up cache directory for Buildx")
				oldCacheBuildxPath = util.LOCAL_BUILDX_LOCATION + "/old"
				localCachePath = util.LOCAL_BUILDX_CACHE_LOCATION
				err = impl.setupCacheForBuildx(ciContext, localCachePath, oldCacheBuildxPath)
				if err != nil {
					return "", err
				}
				oldCacheBuildxPath = oldCacheBuildxPath + "/cache"
			}

			// need to export the cache after the build if k8s driver mode is enabled.
			// when we use k8s driver, if we give export cache flag in the build command itself then all the k8s driver nodes will push the cache to same location.
			// then we will endup with having any one of the node cache in the end and we cannot use this cache for all the platforms in subsequent builds.

			// so we will export the cache after build for all the platforms independently at different locations.
			// refer buildxExportCacheFunc

			multiNodeK8sDriver := useBuildxK8sDriver && len(eligibleK8sDriverNodes) > 1
			exportBuildxCacheAfterBuild := ciRequest.AsyncBuildxCacheExport && multiNodeK8sDriver
			dockerBuild, buildxExportCacheFunc = impl.getBuildxBuildCommand(ciContext, exportBuildxCacheAfterBuild, cacheEnabled, ciRequest.BuildxCacheModeMin, dockerBuild, oldCacheBuildxPath, localCachePath, dest, dockerBuildConfig, dockerfilePath)
		} else {
			dockerBuild = fmt.Sprintf("%s -f %s --network host -t %s %s", dockerBuild, dockerfilePath, ciRequest.DockerRepository, dockerBuildConfig.BuildContext)
		}
		dockerBuildStageMetadata := bean2.DockerBuildStageMetadata{TargetPlatforms: utils.ConvertTargetPlatformStringToObject(ciBuildConfig.DockerBuildConfig.TargetPlatform)}
		buildLogs := []any{"Docker build started..."}
		if envVars.ShowDockerBuildCmdInLogs {
			buildLogs = []any{"Starting docker build : ", dockerBuild}
		}
		err = util.ExecuteWithStageInfoLogWithMetadata(
			util.DOCKER_BUILD,
			dockerBuildStageMetadata,
			impl.buildImageStage(ciContext, dockerBuild, useBuildxK8sDriver, k8sClient, buildLogs),
		)
		if err != nil && !errors.Is(err, BuilderPodDeletedError) {
			return "", err
		} else if errors.Is(err, BuilderPodDeletedError) {
			// Log error message for builder pod interruption due to
			util.LogError(BuilderPodDeletedError)
			util.LogWarn("Frequent spot interruptions can lead to build failures.\n",
				"Consider using a different node type or increasing the spot interruption tolerance.\n")
			maxRetry := ciRequest.BuildxInterruptionMaxRetry - 1 // -1 because we already tried once
			callback := func(retriesLeft int) error {
				attempt := maxRetry - retriesLeft
				reBuildLogs := []any{fmt.Sprintf("Docker re build started (Attempt %d)...", attempt)}
				if envVars.ShowDockerBuildCmdInLogs {
					reBuildLogs = []any{fmt.Sprintf("Starting re docker build (Attempt %d) : ", attempt), dockerBuild}
				}
				return impl.executeDockerReBuild(ciContext, k8sClient, useBuildxK8sDriver, dockerBuild,
					deploymentNames, dockerBuildStageMetadata, reBuildLogs)
			}
			err = retryFunc.RetryWithOutLogging(callback, retryFunc.IsRetryableError, maxRetry, 1*time.Second)
			if err != nil {
				log.Println(util.DEVTRON, " error in executing docker re-build ", " err: ", err)
				return "", err
			}
		}

		if buildxExportCacheFunc != nil {
			//todo do we need to throw error here? discuss
			util.ExecuteWithStageInfoLog(util.EXPORT_BUILD_CACHE, buildxExportCacheFunc)
		}

		if useBuildK8sDriver, eligibleK8sDriverNodes := dockerBuildConfig.CheckForBuildXK8sDriver(); useBuildK8sDriver {

			buildxCleanupSatge := func() error {
				err = impl.CleanBuildxK8sDriver(ciContext, eligibleK8sDriverNodes)
				if err != nil {
					log.Println(util.DEVTRON, " error in cleaning buildx K8s driver ", " err: ", err)
				}
				return nil
			}

			// do not need to handle the below error
			util.ExecuteWithStageInfoLog(util.CLEANUP_BUILDX_BUILDER, buildxCleanupSatge)
		}

		if !useBuildx {
			err = impl.tagDockerBuild(ciContext, ciRequest.DockerRepository, dest)
			if err != nil {
				return "", err
			}
		} else {
			return dest, nil
		}
	} else if ciBuildConfig.CiBuildType == BUILDPACK_BUILD_TYPE {

		buildPacksImageBuildStage := func() error {
			buildPackParams := ciRequest.CiBuildConfig.BuildPackConfig
			projectPath := buildPackParams.ProjectPath
			if projectPath == "" || !strings.HasPrefix(projectPath, "./") {
				projectPath = "./" + projectPath
			}
			impl.handleLanguageVersion(ciContext, projectPath, buildPackParams)
			buildPackCmd := fmt.Sprintf("pack build %s --path %s --builder %s", dest, projectPath, buildPackParams.BuilderId)
			BuildPackArgsMap := buildPackParams.Args
			for k, v := range BuildPackArgsMap {
				buildPackCmd = buildPackCmd + " --env " + k + "=" + v
			}

			if len(buildPackParams.BuildPacks) > 0 {
				for _, buildPack := range buildPackParams.BuildPacks {
					buildPackCmd = buildPackCmd + " --buildpack " + buildPack
				}
			}
			log.Println(" -----> " + buildPackCmd)
			err = impl.executeCmd(ciContext, buildPackCmd)
			if err != nil {
				return err
			}
			builderRmCmdString := "docker image rm " + buildPackParams.BuilderId
			builderRmCmd := impl.GetCommandToExecute(builderRmCmdString)
			err := builderRmCmd.Run()
			if err != nil {
				return err
			}
			return nil
		}

		if err = util.ExecuteWithStageInfoLog(util.BUILD_PACK_BUILD, buildPacksImageBuildStage); err != nil {
			return "", err
		}

	}

	return dest, nil
}

func (impl *DockerHelperImpl) runDockerBuildCommand(ciContext cicxt.CiContext, dockerBuild string) error {
	errChan := make(chan error)
	go func() {
		errChan <- impl.executeCmdWithCtx(ciContext, dockerBuild)
	}()
	select {
	case <-ciContext.Done():
		return ciContext.Err()
	case err := <-errChan:
		return err
	}
}

func getDockerBuildFlagsMap(dockerBuildConfig *DockerBuildConfig) map[string]string {
	dockerBuildFlags := make(map[string]string)
	dockerBuildArgsMap := dockerBuildConfig.Args
	for k, v := range dockerBuildArgsMap {
		flagKey := fmt.Sprintf("%s %s", BUILD_ARG_FLAG, k)
		dockerBuildFlags[flagKey] = parseDockerFlagParam(v)
	}
	dockerBuildOptionsMap := dockerBuildConfig.DockerBuildOptions
	for k, v := range dockerBuildOptionsMap {
		flagKey := "--" + k
		dockerBuildFlags[flagKey] = parseDockerFlagParam(v)
	}
	return dockerBuildFlags
}

func parseDockerFlagParam(param string) string {
	value := param
	if strings.HasPrefix(param, DEVTRON_ENV_VAR_PREFIX) {
		value = os.Getenv(strings.TrimPrefix(param, DEVTRON_ENV_VAR_PREFIX))
	}

	return wrapSingleOrDoubleQuotedValue(value)
}

func wrapSingleOrDoubleQuotedValue(value string) string {
	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
		unquotedString := strings.Trim(value, `'`)
		return fmt.Sprintf(`='%s'`, unquotedString)
	}

	unquotedString := strings.Trim(value, `"`)

	return fmt.Sprintf(`="%s"`, unquotedString)
}

func getDockerfilePath(CiBuildConfig *CiBuildConfigBean, checkoutPath string) string {
	var dockerFilePath string
	if CiBuildConfig.CiBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerFilePath = GetSelfManagedDockerfilePath(checkoutPath)
	} else {
		dockerFilePath = CiBuildConfig.DockerBuildConfig.DockerfilePath
	}
	return dockerFilePath
}

// getBuildxExportCacheFunc  will concurrently execute the given export cache commands
func (impl *DockerHelperImpl) getBuildxExportCacheFunc(ciContext cicxt.CiContext, exportCacheCmds map[string]string) func() error {
	exportCacheFunc := func() error {
		// run export cache cmd for buildx
		if len(exportCacheCmds) > 0 {
			log.Println("exporting build caches...")
			wg := sync.WaitGroup{}
			wg.Add(len(exportCacheCmds))
			for platform, exportCacheCmd := range exportCacheCmds {
				go func(platform, exportCacheCmd string) {
					log.Println("exporting build cache, platform : ", platform)
					log.Println(exportCacheCmd)
					err := impl.executeCmd(ciContext, exportCacheCmd)
					if err != nil {
						log.Println("error in exporting", "err:", err)
						//not returning as need to mark waitGroup done
					}
					wg.Done()
				}(platform, exportCacheCmd)
			}
			wg.Wait()
		}
		return nil
	}
	return exportCacheFunc
}

// getExportCacheCmds will return build commands exclusively for exporting cache for all the given target platforms.
func getExportCacheCmds(targetPlatforms, dockerBuild, localCachePath string, useCacheMin bool) map[string]string {

	cacheMode := CacheModeMax
	if useCacheMin {
		cacheMode = CacheModeMin
	}

	cacheCmd := "%s --platform=%s --cache-to=type=local,dest=%s,mode=" + cacheMode
	platforms := utils.ConvertTargetPlatformStringToList(targetPlatforms)

	exportCacheCmds := make(map[string]string)
	for _, platform := range platforms {
		cachePath := strings.Join(strings.Split(platform, "/"), "-")
		exportCacheCmds[platform] = fmt.Sprintf(cacheCmd, dockerBuild, platform, localCachePath+"/"+cachePath)
	}
	return exportCacheCmds
}

func getSourceCaches(targetPlatforms, oldCachePathLocation string) string {
	cacheCmd := " --cache-from=type=local,src=%s "
	platforms := strings.Split(targetPlatforms, ",")
	allCachePaths := make([]string, 0, len(platforms))
	for _, platform := range platforms {
		cachePath := strings.Join(strings.Split(platform, "/"), "-")
		allCachePaths = append(allCachePaths, fmt.Sprintf(cacheCmd, oldCachePathLocation+"/"+cachePath))
	}
	return strings.Join(allCachePaths, " ")
}

func (impl *DockerHelperImpl) getBuildxBuildCommandV2(ciContext cicxt.CiContext, cacheEnabled bool, useCacheMin bool, dockerBuild, oldCacheBuildxPath, localCachePath, dest string, dockerBuildConfig *DockerBuildConfig, dockerfilePath string) (string, func() error) {
	dockerBuild = fmt.Sprintf("%s %s -f %s --network host --allow network.host --allow security.insecure", dockerBuild, dockerBuildConfig.BuildContext, dockerfilePath)
	exportCacheCmds := make(map[string]string)

	provenanceFlag := dockerBuildConfig.GetProvenanceFlag()
	dockerBuild = fmt.Sprintf("%s %s", dockerBuild, provenanceFlag)

	// separate out export cache and source cache cmds here
	isTargetPlatformSet := dockerBuildConfig.TargetPlatform != ""
	if isTargetPlatformSet {
		if cacheEnabled {
			exportCacheCmds = getExportCacheCmds(dockerBuildConfig.TargetPlatform, dockerBuild, localCachePath, useCacheMin)
		}

		dockerBuild = fmt.Sprintf("%s --platform %s", dockerBuild, dockerBuildConfig.TargetPlatform)
	}

	if cacheEnabled {
		dockerBuild = fmt.Sprintf("%s %s", dockerBuild, getSourceCaches(dockerBuildConfig.TargetPlatform, oldCacheBuildxPath))
	}

	manifestLocation := util.LOCAL_BUILDX_LOCATION + "/manifest.json"
	dockerBuild = fmt.Sprintf("%s -t %s --push --metadata-file %s", dockerBuild, dest, manifestLocation)

	return dockerBuild, impl.getBuildxExportCacheFunc(ciContext, exportCacheCmds)
}

func (impl *DockerHelperImpl) getBuildxBuildCommandV1(cacheEnabled bool, useCacheMin bool, dockerBuild, oldCacheBuildxPath, localCachePath, dest string, dockerBuildConfig *DockerBuildConfig, dockerfilePath string) (string, func() error) {

	cacheMode := CacheModeMax
	if useCacheMin {
		cacheMode = CacheModeMin
	}
	dockerBuild = fmt.Sprintf("%s -f %s -t %s --push %s --network host --allow network.host --allow security.insecure", dockerBuild, dockerfilePath, dest, dockerBuildConfig.BuildContext)
	if cacheEnabled {
		dockerBuild = fmt.Sprintf("%s --cache-to=type=local,dest=%s,mode=%s --cache-from=type=local,src=%s", dockerBuild, localCachePath, cacheMode, oldCacheBuildxPath)
	}

	isTargetPlatformSet := dockerBuildConfig.TargetPlatform != ""
	if isTargetPlatformSet {
		dockerBuild = fmt.Sprintf("%s --platform %s", dockerBuild, dockerBuildConfig.TargetPlatform)
	}

	provenanceFlag := dockerBuildConfig.GetProvenanceFlag()
	dockerBuild = fmt.Sprintf("%s %s", dockerBuild, provenanceFlag)
	manifestLocation := util.LOCAL_BUILDX_LOCATION + "/manifest.json"
	dockerBuild = fmt.Sprintf("%s --metadata-file %s", dockerBuild, manifestLocation)

	return dockerBuild, nil
}

func (impl *DockerHelperImpl) getBuildxBuildCommand(ciContext cicxt.CiContext, exportBuildxCacheAfterBuild bool, cacheEnabled bool, useCacheMin bool, dockerBuild, oldCacheBuildxPath, localCachePath, dest string, dockerBuildConfig *DockerBuildConfig, dockerfilePath string) (string, func() error) {
	if exportBuildxCacheAfterBuild {
		return impl.getBuildxBuildCommandV2(ciContext, cacheEnabled, useCacheMin, dockerBuild, oldCacheBuildxPath, localCachePath, dest, dockerBuildConfig, dockerfilePath)
	}
	return impl.getBuildxBuildCommandV1(cacheEnabled, useCacheMin, dockerBuild, oldCacheBuildxPath, localCachePath, dest, dockerBuildConfig, dockerfilePath)
}

func (impl *DockerHelperImpl) handleLanguageVersion(ciContext cicxt.CiContext, projectPath string, buildpackConfig *BuildPackConfig) {
	fileData, err := os.ReadFile("/buildpack.json")
	if err != nil {
		log.Println("error occurred while reading buildpack json", err)
		return
	}
	var buildpackDataArray []*BuildpackVersionConfig
	err = json.Unmarshal(fileData, &buildpackDataArray)
	if err != nil {
		log.Println("error occurred while reading buildpack json", string(fileData))
		return
	}
	language := buildpackConfig.Language
	// languageVersion := buildpackConfig.LanguageVersion
	buildpackEnvArgs := buildpackConfig.Args
	languageVersion, present := buildpackEnvArgs["DEVTRON_LANG_VERSION"]
	if !present {
		return
	}
	var matchedBuildpackConfig *BuildpackVersionConfig
	for _, versionConfig := range buildpackDataArray {
		builderPrefix := versionConfig.BuilderPrefix
		configLanguage := versionConfig.Language
		builderId := buildpackConfig.BuilderId
		if strings.HasPrefix(builderId, builderPrefix) && strings.ToLower(language) == configLanguage {
			matchedBuildpackConfig = versionConfig
			break
		}
	}
	if matchedBuildpackConfig != nil {
		fileName := matchedBuildpackConfig.FileName
		finalPath := filepath.Join(projectPath, "./"+fileName)
		_, err := os.Stat(finalPath)
		fileNotExists := errors.Is(err, os.ErrNotExist)
		if fileNotExists {
			file, err := os.Create(finalPath)
			if err != nil {
				fmt.Println("error occurred while creating file at path " + finalPath)
				return
			}
			entryRegex := matchedBuildpackConfig.EntryRegex
			languageEntry := fmt.Sprintf(entryRegex, languageVersion)
			_, err = file.WriteString(languageEntry)
			log.Println(util.DEVTRON, fmt.Sprintf(" file %s created for language %s with version %s", finalPath, language, languageVersion))
		} else if matchedBuildpackConfig.FileOverride {
			log.Println("final Path is ", finalPath)
			ext := filepath.Ext(finalPath)
			if ext == ".json" {
				jqCmd := fmt.Sprintf("jq '.engines.node' %s", finalPath)
				outputBytes, err := exec.Command("/bin/sh", "-c", jqCmd).Output()
				if err != nil {
					log.Println("error occurred while fetching node version", "err", err)
					return
				}
				if strings.TrimSpace(string(outputBytes)) == "null" {
					tmpJsonFile := "./tmp.json"
					versionUpdateCmd := fmt.Sprintf("jq '.engines.node = \"%s\"' %s >%s", languageVersion, finalPath, tmpJsonFile)
					err := impl.executeCmd(ciContext, versionUpdateCmd)
					if err != nil {
						log.Println("error occurred while inserting node version", "err", err)
						return
					}
					fileReplaceCmd := fmt.Sprintf("mv %s %s", tmpJsonFile, finalPath)
					err = impl.executeCmd(ciContext, fileReplaceCmd)
					if err != nil {
						log.Println("error occurred while executing cmd ", fileReplaceCmd, "err", err)
						return
					}
				}
			}
		} else {
			log.Println("file already exists, so ignoring version override!!", finalPath)
		}
	}

}

func (impl *DockerHelperImpl) executeCmd(ciContext cicxt.CiContext, cmd string) error {
	exeCmd := impl.GetCommandToExecute(cmd)
	err := impl.cmdExecutor.RunCommand(ciContext, exeCmd)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (impl *DockerHelperImpl) executeCmdWithCtx(ciContext cicxt.CiContext, cmd string) error {
	exeCmd := impl.GetCommandToExecute(cmd)
	err := impl.cmdExecutor.RunCommandWithCtx(ciContext, exeCmd)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (impl *DockerHelperImpl) tagDockerBuild(ciContext cicxt.CiContext, dockerRepository string, dest string) error {
	dockerTag := "docker tag " + dockerRepository + ":latest" + " " + dest
	log.Println(" -----> " + dockerTag)
	dockerTagCMD := impl.GetCommandToExecute(dockerTag)
	err := impl.cmdExecutor.RunCommand(ciContext, dockerTagCMD)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (impl *DockerHelperImpl) setupCacheForBuildx(ciContext cicxt.CiContext, localCachePath string, oldCacheBuildxPath string) error {
	err := impl.checkAndCreateDirectory(ciContext, localCachePath)
	if err != nil {
		return err
	}
	err = impl.checkAndCreateDirectory(ciContext, oldCacheBuildxPath)
	if err != nil {
		return err
	}
	copyContent := "cp -R " + localCachePath + " " + oldCacheBuildxPath
	copyContentCmd := exec.Command("/bin/sh", "-c", copyContent)
	err = impl.cmdExecutor.RunCommand(ciContext, copyContentCmd)

	if err != nil {
		log.Println(err)
		return err
	}

	cleanContent := "rm -rf " + localCachePath + "/*"
	cleanContentCmd := exec.Command("/bin/sh", "-c", cleanContent)
	err = impl.cmdExecutor.RunCommand(ciContext, cleanContentCmd)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (impl *DockerHelperImpl) createBuildxBuilder(ciContext cicxt.CiContext, dockerConnection, buildxDriverImage string) error {
	buildkitToml := ""
	if dockerConnection == util.SECUREWITHCERT {
		buildkitToml = fmt.Sprintf("--config %s", BuildkitdConfigPath)
	}
	multiPlatformCmd := fmt.Sprintf("docker buildx create --use --buildkitd-flags '--allow-insecure-entitlement network.host --allow-insecure-entitlement security.insecure' %s", buildkitToml)
	driverOptions := getBuildXDriverOptionsWithImage(buildxDriverImage, "")
	if len(driverOptions) > 0 {
		multiPlatformCmd += " '--driver-opt=%s' "
		multiPlatformCmd = fmt.Sprintf(multiPlatformCmd, driverOptions)
	}

	log.Println(" -----> " + multiPlatformCmd)
	dockerBuildCMD := impl.GetCommandToExecute(multiPlatformCmd)
	err := impl.cmdExecutor.RunCommand(ciContext, dockerBuildCMD)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (impl *DockerHelperImpl) installAllSupportedPlatforms(ciContext cicxt.CiContext) error {
	multiPlatformCmd := "docker run --privileged --rm quay.io/devtron/binfmt:stable --install all"
	log.Println(" -----> " + multiPlatformCmd)
	dockerBuildCMD := impl.GetCommandToExecute(multiPlatformCmd)
	err := impl.cmdExecutor.RunCommand(ciContext, dockerBuildCMD)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (impl *DockerHelperImpl) checkAndCreateDirectory(ciContext cicxt.CiContext, localCachePath string) error {
	makeDirCmd := "mkdir -p " + localCachePath
	pathCreateCommand := exec.Command("/bin/sh", "-c", makeDirCmd)
	err := impl.cmdExecutor.RunCommand(ciContext, pathCreateCommand)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func BuildDockerImagePath(ciRequest *CommonWorkflowRequest) (string, error) {
	return utils.BuildDockerImagePath(bean.DockerRegistryInfo{
		DockerImageTag:     ciRequest.DockerImageTag,
		DockerRegistryId:   ciRequest.DockerRegistryId,
		DockerRegistryType: ciRequest.DockerRegistryType,
		DockerRegistryURL:  ciRequest.IntermediateDockerRegistryUrl,
		DockerRepository:   ciRequest.DockerRepository,
	})

}

func (impl *DockerHelperImpl) PushArtifact(ciContext cicxt.CiContext, dest string) error {
	//awsLogin := "$(aws ecr get-login --no-include-email --region " + ciRequest.AwsRegion + ")"
	dockerPush := "docker push " + dest
	log.Println("-----> " + dockerPush)
	dockerPushCMD := impl.GetCommandToExecute(dockerPush)
	err := impl.cmdExecutor.RunCommand(ciContext, dockerPushCMD)
	if err != nil {
		log.Println(err)
		return err
	}

	// digest := extractDigestUsingPull(dest)
	// log.Println("Digest -----> ", digest)
	// return digest, nil
	return nil
}

func (impl *DockerHelperImpl) ExtractDigestForBuildx(dest string, ciRequest *CommonWorkflowRequest) (string, error) {

	var digest string
	var err error
	manifestLocation := util.LOCAL_BUILDX_LOCATION + "/manifest.json"
	digest, err = readImageDigestFromManifest(manifestLocation)
	if err != nil {
		log.Println("error occurred while extracting digest from manifest reason ", err)
		err = nil // would extract digest using docker pull cmd
	}
	if digest == "" {
		dockerAuthConfig := impl.GetDockerAuthConfigForPrivateRegistries(ciRequest)
		// if UseDockerApiToGetDigest is true then fetches digest from docker api else uses docker pull command and then parse the result
		digest, err = impl.ExtractDigestFromImage(dest, ciRequest.UseDockerApiToGetDigest, dockerAuthConfig)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error in extracting digest from image %s, err:", dest), err)
		}
	}
	log.Println("Digest -----> ", digest)

	return digest, err
}

func (impl *DockerHelperImpl) ExtractDigestFromImage(image string, useDockerApiToGetDigest bool, dockerAuthConfig *bean.DockerAuthConfig) (string, error) {
	var digest string
	var err error
	if useDockerApiToGetDigest {
		log.Println("fetching digest from docker api")
		digest, err = dockerOperations.GetImageDigestByImage(context.Background(), image, dockerAuthConfig)
		if err != nil {
			fmt.Println(fmt.Sprintf("get digest via docker api error, error in extracting digest from image %s, err:", image), err)
			return "", err
		}
	} else {
		log.Println("fetching digest using docker pull command")
		digest, err = impl.ExtractDigestUsingPull(image)
		if err != nil {
			fmt.Println(fmt.Sprintf("docker pull image error, error in extracting digest from image %s, err:", image), err)
			return "", err
		}
	}
	return digest, nil
}

func (impl *DockerHelperImpl) ExtractDigestUsingPull(dest string) (string, error) {
	dockerPull := "docker pull " + dest
	dockerPullCmd := impl.GetCommandToExecute(dockerPull)
	digest, err := runGetDockerImageDigest(dockerPullCmd)
	if err != nil {
		log.Println(err)
	}
	return digest, err
}

func runGetDockerImageDigest(cmd *exec.Cmd) (string, error) {
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd.Stdout = mw
	cmd.Stderr = mw
	if err := cmd.Run(); err != nil {
		return "", err
	}
	output := stdBuffer.String()
	outArr := strings.Split(output, "\n")
	var digest string
	for _, s := range outArr {
		if strings.HasPrefix(s, "Digest: ") {
			digest = s[strings.Index(s, "sha256:"):]
		}

	}
	return digest, nil
}

func readImageDigestFromManifest(manifestFilePath string) (string, error) {
	manifestFile, err := ioutil.ReadFile(manifestFilePath)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	err = json.Unmarshal(manifestFile, &data)
	if err != nil {
		return "", err
	}
	imageDigest, found := data["containerimage.digest"]
	if !found {
		return "", nil
	}
	return imageDigest.(string), nil
}

func (impl *DockerHelperImpl) createBuildxBuilderForMultiArchBuild(ciContext cicxt.CiContext, dockerConnection, buildxDriverImage string) error {
	err := impl.installAllSupportedPlatforms(ciContext)
	if err != nil {
		return err
	}
	err = impl.createBuildxBuilder(ciContext, dockerConnection, buildxDriverImage)
	if err != nil {
		return err
	}
	return nil
}

func (impl *DockerHelperImpl) createBuildxBuilderWithK8sDriver(ciContext cicxt.CiContext, propagateLabelsInBuildxPod bool, dockerConnection, buildxDriverImage string, builderNodes []map[string]string, ciPipelineId, ciWorkflowId int) ([]string, error) {
	deploymentNames := make([]string, 0)
	if len(builderNodes) == 0 {
		return deploymentNames, errors.New("atleast one node is expected for builder with kubernetes driver")
	}
	for i := 0; i < len(builderNodes); i++ {
		nodeOpts := builderNodes[i]
		builderCmd, deploymentName, err := getBuildxK8sDriverCmd(propagateLabelsInBuildxPod, dockerConnection, buildxDriverImage, nodeOpts, ciPipelineId, ciWorkflowId)
		if err != nil {
			return deploymentNames, err
		}
		deploymentNames = append(deploymentNames, deploymentName)
		// first node is used as default node, we create builder with --use flag, then we append other nodes
		if i == 0 {
			builderCmd = fmt.Sprintf("%s %s", builderCmd, "--use")
		} else {
			// appending other nodes to the builder,except default node ,since we already added it
			builderCmd = fmt.Sprintf("%s %s", builderCmd, "--append")
		}

		fmt.Println(util.DEVTRON, " cmd : ", builderCmd)
		builderExecCmd := impl.GetCommandToExecute(builderCmd)
		err = impl.cmdExecutor.RunCommand(ciContext, builderExecCmd)
		if err != nil {
			fmt.Println(util.DEVTRON, " builderCmd : ", builderCmd, " err : ", err)
			return deploymentNames, err
		}
	}
	return deploymentNames, nil
}

func (impl *DockerHelperImpl) CleanBuildxK8sDriver(ciContext cicxt.CiContext, nodes []map[string]string) error {
	nodeNames := make([]string, 0)
	for _, nOptsMp := range nodes {
		if nodeName, ok := nOptsMp["node"]; ok && nodeName != "" {
			nodeNames = append(nodeNames, nodeName)
		}
	}
	err := impl.leaveNodesFromBuildxK8sDriver(ciContext, nodeNames)
	if err != nil {
		log.Println(util.DEVTRON, " error in deleting nodes created by ci-runner , err : ", err)
		return err
	}
	log.Println(util.DEVTRON, "successfully cleaned up buildx k8s driver")
	return nil
}

func (impl *DockerHelperImpl) leaveNodesFromBuildxK8sDriver(ciContext cicxt.CiContext, nodeNames []string) error {
	var err error
	for _, node := range nodeNames {
		createCmd := fmt.Sprintf("docker buildx create --name=%s --node=%s --leave", BUILDX_K8S_DRIVER_NAME, node)
		fmt.Println(util.DEVTRON, " cmd : ", createCmd)
		execCreateCmd := impl.GetCommandToExecute(createCmd)
		err = impl.cmdExecutor.RunCommand(ciContext, execCreateCmd)
		if err != nil {
			log.Println(util.DEVTRON, "error in leaving node : ", err)
			break
		}
	}
	impl.removeBuildxDriver(ciContext) //driver cleanup
	return err
}

func (impl *DockerHelperImpl) removeBuildxDriver(ciContext cicxt.CiContext) {
	removeCmd := fmt.Sprintf("docker buildx rm %s", BUILDX_K8S_DRIVER_NAME)
	fmt.Println(util.DEVTRON, " cmd : ", removeCmd)
	execRemoveCmd := impl.GetCommandToExecute(removeCmd)
	err := impl.cmdExecutor.RunCommand(ciContext, execRemoveCmd)
	if err != nil {
		log.Println("error in executing docker buildx remove command", "err", err)
		//not returning error here as this is just a cleanup job, not making it blocking
	}
}

// this function is deprecated, use cmdExecutor.RunCommand instead
func (impl *DockerHelperImpl) runCmd(cmd string) (error, *bytes.Buffer) {
	fmt.Println(util.DEVTRON, " cmd : ", cmd)
	builderCreateCmd := impl.GetCommandToExecute(cmd)
	errBuf := &bytes.Buffer{}
	builderCreateCmd.Stderr = errBuf
	err := builderCreateCmd.Run()
	return err, errBuf
}

func getBuildxK8sDriverCmd(propagateLabelsInBuildxPod bool, dockerConnection, buildxDriverImage string, driverOpts map[string]string, ciPipelineId, ciWorkflowId int) (string, string, error) {
	buildxCreate := "docker buildx create --buildkitd-flags '--allow-insecure-entitlement network.host --allow-insecure-entitlement security.insecure' --name=%s --driver=kubernetes --node=%s --bootstrap "
	nodeName := driverOpts["node"]
	if nodeName == "" {
		nodeName = BUILDX_NODE_NAME + fmt.Sprintf("%v-%v-", ciPipelineId, ciWorkflowId) + util.Generate(3) // need this to generate unique name for builder node in same builder.
	}
	buildxCreate = fmt.Sprintf(buildxCreate, BUILDX_K8S_DRIVER_NAME, nodeName)
	platforms := driverOpts["platform"]
	if platforms != "" {
		buildxCreate += " --platform=%s "
		buildxCreate = fmt.Sprintf(buildxCreate, platforms)
	}
	// add driver options for app labels and annotations
	var err error
	if propagateLabelsInBuildxPod {
		driverOpts["driverOptions"], err = getBuildXDriverOptionsWithLabelsAndAnnotations(driverOpts["driverOptions"])
		if err != nil {
			return "", "", err
		}
	}

	driverOpts["driverOptions"] = getBuildXDriverOptionsWithImage(buildxDriverImage, driverOpts["driverOptions"])
	if len(driverOpts["driverOptions"]) > 0 {
		buildxCreate += " '--driver-opt=%s' "
		buildxCreate = fmt.Sprintf(buildxCreate, driverOpts["driverOptions"])
	}
	buildkitToml := ""
	if dockerConnection == util.SECUREWITHCERT {
		buildkitToml = fmt.Sprintf("--config %s", BuildkitdConfigPath)
	}
	buildxCreate = fmt.Sprintf("%s %s", buildxCreate, buildkitToml)
	return buildxCreate, nodeName, nil
}

func getBuildXDriverOptionsWithImage(buildxDriverImage, driverOptions string) string {
	if len(buildxDriverImage) > 0 {
		driverImageOption := fmt.Sprintf("\"image=%s\"", buildxDriverImage)
		if len(driverOptions) > 0 {
			driverOptions += fmt.Sprintf(",%s", driverImageOption)
		} else {
			driverOptions = driverImageOption
		}
	}
	return driverOptions
}

func getBuildXDriverOptionsWithLabelsAndAnnotations(driverOptions string) (string, error) {
	// not passing annotation as of now because --driver-opt=annotations is not supported by buildx if contains quotes
	labels := make(map[string]string)

	// Read labels from file
	labelsPath := utils.DEVTRON_SELF_DOWNWARD_API_VOLUME_PATH + "/" + utils.POD_LABELS
	labelsOut, err := readFileAndLogErrors(labelsPath)
	if err != nil {
		return "", err
	}

	// Parse labels
	if len(labelsOut) > 0 {
		labels = parseKeyValuePairs(string(labelsOut))
	}

	// Combine driver options
	driverOptions = getBuildXDriverOptions(utils.POD_LABELS, labels, driverOptions)

	//annotations := make(map[string]string)
	//annotationsPath := utils.DEVTRON_SELF_DOWNWARD_API_VOLUME_PATH + "/" + utils.POD_ANNOTATIONS
	//annotationsOut, err := readFileAndLogErrors(annotationsPath)
	//if err != nil {
	//	return "", err
	//}
	//if len(annotationsOut) > 0 {
	//	annotations = parseKeyValuePairs(string(annotationsOut))
	//}
	//driverOptions = getBuildXDriverOptions(utils.POD_ANNOTATIONS, annotations, driverOptions)

	return driverOptions, nil
}

func readFileAndLogErrors(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println(util.DEVTRON, "file not found at path:", filePath)
			return content, nil
		} else {
			log.Println(util.DEVTRON, "IO error while reading file at path:", filePath, "err:", err)
			return nil, err
		}
	}
	return content, nil
}

func parseKeyValuePairs(input string) map[string]string {
	keyValuePairs := make(map[string]string)
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
				keyValuePairs[key] = value
			}
		}
	}
	return keyValuePairs
}

func getBuildXDriverOptions(optionType string, options map[string]string, driverOptions string) string {
	if len(options) > 0 {
		optionStr := fmt.Sprintf("\"%s=", optionType)
		for k, v := range options {
			optionStr += fmt.Sprintf("%s=%s,", k, v)
		}
		optionStr = strings.TrimSuffix(optionStr, ",")
		optionStr += "\""

		if len(driverOptions) > 0 {
			driverOptions += fmt.Sprintf(",%s", optionStr)
		} else {
			driverOptions = optionStr
		}
	}
	return driverOptions
}

func (impl *DockerHelperImpl) StopDocker(ciContext cicxt.CiContext) error {
	cmd := exec.Command("docker", "ps", "-a", "-q")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		stopCmdS := "docker stop -t 5 $(docker ps -a -q)"
		log.Println(util.DEVTRON, " -----> stopping docker container")
		stopCmd := impl.GetCommandToExecute(stopCmdS)
		err := impl.cmdExecutor.RunCommand(ciContext, stopCmd)
		log.Println(util.DEVTRON, " -----> stopped docker container")
		if err != nil {
			log.Fatal(err)
			return err
		}
		removeContainerCmds := "docker rm -v -f $(docker ps -a -q)"
		log.Println(util.DEVTRON, " -----> removing docker container")
		removeContainerCmd := impl.GetCommandToExecute(removeContainerCmds)
		err = impl.cmdExecutor.RunCommand(ciContext, removeContainerCmd)
		log.Println(util.DEVTRON, " -----> removed docker container")
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	file := "/var/run/docker.pid"
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		return err
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Println(err)
		return err
	}
	// Kill the process
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(util.DEVTRON, " -----> checking docker status")
	impl.DockerdUpCheck() // FIXME: this call should be removed
	// ensureDockerDaemonHasStopped(20)
	return nil
}

func (impl *DockerHelperImpl) ensureDockerDaemonHasStopped(retryCount int) error {
	var err error
	retry := 0
	for err == nil {
		time.Sleep(1 * time.Second)
		err = impl.DockerdUpCheck()
		retry++
		if retry == retryCount {
			break
		}
	}
	return err
}

func (impl *DockerHelperImpl) waitForDockerDaemon(retryCount int) error {
	err := impl.DockerdUpCheck()
	retry := 0
	for err != nil {
		if retry == retryCount {
			break
		}
		time.Sleep(1 * time.Second)
		err = impl.DockerdUpCheck()
		retry++
	}
	return err
}

func (impl *DockerHelperImpl) DockerdUpCheck() error {
	dockerCheck := "docker ps"
	dockerCheckCmd := impl.GetCommandToExecute(dockerCheck)
	err := dockerCheckCmd.Run()
	return err
}

func ValidBuildxK8sDriverOptions(ciRequest *CommonWorkflowRequest) (bool, []map[string]string) {
	valid := ciRequest != nil && ciRequest.CiBuildConfig != nil && ciRequest.CiBuildConfig.DockerBuildConfig != nil
	if valid {
		return ciRequest.CiBuildConfig.DockerBuildConfig.CheckForBuildXK8sDriver()
	}
	return false, nil
}

func GetSelfManagedDockerfilePath(checkoutPath string) string {
	return filepath.Join(util.WORKINGDIR, checkoutPath, "./Dockerfile")
}

func (impl *DockerHelperImpl) GetDockerAuthConfigForPrivateRegistries(workflowRequest *CommonWorkflowRequest) *bean.DockerAuthConfig {
	var dockerAuthConfig *bean.DockerAuthConfig
	switch workflowRequest.DockerRegistryType {
	case REGISTRY_TYPE_GCR:
		if len(workflowRequest.DockerPassword) > 0 {
			dockerAuthConfig = &bean.DockerAuthConfig{
				RegistryType:          bean.RegistryTypeGcr,
				CredentialFileJsonGcr: workflowRequest.DockerPassword,
				IsRegistryPrivate:     true,
			}
		}
	case DOCKER_REGISTRY_TYPE_ECR:
		if len(workflowRequest.AccessKey) > 0 && len(workflowRequest.SecretKey) > 0 && len(workflowRequest.AwsRegion) > 0 {
			dockerAuthConfig = &bean.DockerAuthConfig{
				RegistryType:       bean.RegistryTypeEcr,
				AccessKeyEcr:       workflowRequest.AccessKey,
				SecretAccessKeyEcr: workflowRequest.SecretKey,
				EcrRegion:          workflowRequest.AwsRegion,
				IsRegistryPrivate:  true,
			}
		}
	default:
		if len(workflowRequest.DockerUsername) > 0 && len(workflowRequest.DockerPassword) > 0 {
			dockerAuthConfig = &bean.DockerAuthConfig{
				Username:          workflowRequest.DockerUsername,
				Password:          workflowRequest.DockerPassword,
				IsRegistryPrivate: true,
			}
		}
	}
	return dockerAuthConfig
}
