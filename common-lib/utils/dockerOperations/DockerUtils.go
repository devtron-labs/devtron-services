package dockerOperations

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"strings"
)

// LoadGcrCredentials loads Google service account credentials from a JSON file
func LoadGcrCredentials(credsJson string) (string, string, error) {
	conf, err := google.JWTConfigFromJSON([]byte(credsJson), bean.GcrRegistryScope)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse credentials JSON: %v", err)
	}
	tokenSource := conf.TokenSource(context.Background())
	token, err := tokenSource.Token()
	if err != nil {
		return "", "", fmt.Errorf("failed to obtain token: %v", err)
	}
	return bean.GcrRegistryUsername, token.AccessToken, nil

}

func LoadEcrCredentials(ecrRegion, accessKeyEcr, secretAccessKeyEcr, assumeRoleArn string) (string, string, error) {
	var username, password string
	awsCfg := &aws.Config{
		Region:      aws.String(ecrRegion),
		Credentials: credentials.NewStaticCredentials(accessKeyEcr, secretAccessKeyEcr, ""),
	}
	sess := session.Must(session.NewSession(awsCfg))

	// If an assume role ARN is provided, use STS to assume the cross-account role
	if len(assumeRoleArn) > 0 {
		stsClient := sts.New(sess)
		assumeOutput, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
			RoleArn:         aws.String(assumeRoleArn),
			RoleSessionName: aws.String("devtron-ecr-cross-account"),
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to assume role %s for ecr: %v", assumeRoleArn, err)
		}
		assumedCreds := credentials.NewStaticCredentials(
			*assumeOutput.Credentials.AccessKeyId,
			*assumeOutput.Credentials.SecretAccessKey,
			*assumeOutput.Credentials.SessionToken,
		)
		sess = session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(ecrRegion),
			Credentials: assumedCreds,
		}))
	}

	svc := ecr.New(sess)
	authData, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch Authorization token from ecr: %v", err)
	}
	// ecr returns authToken in base64 decoded format seperated via colon(:) where on the left side of it is the username and on the right is actual token i.e. password
	decodedEcrToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode Authorization token from ecr: %v", err)
	}
	creds := strings.Split(string(decodedEcrToken), ":")
	if len(decodedEcrToken) > 0 && len(creds) > 1 {
		username = creds[0]
		password = creds[1]
	}
	return username, password, nil
}

func getEncodedRegistryAuthForPrivateRegistry(dockerAuth *bean.DockerAuthConfig) (string, error) {
	switch dockerAuth.RegistryType {
	case bean.RegistryTypeEcr:
		// for ecr we get username and password via region, access and secret access tokens
		ecrUsername, ecrPassword, err := LoadEcrCredentials(dockerAuth.EcrRegion, dockerAuth.AccessKeyEcr, dockerAuth.SecretAccessKeyEcr, dockerAuth.AssumeRoleArnEcr)
		if err != nil {
			logrus.Error("error in getting ecr credentials", "err", err)
			return "", err
		}
		dockerAuth.Username = ecrUsername
		dockerAuth.Password = ecrPassword
	case bean.RegistryTypeGcr:
		// for gcr we get username and password via Google creds json
		gcrUsername, gcrPassword, err := LoadGcrCredentials(dockerAuth.CredentialFileJsonGcr)
		if err != nil {
			logrus.Error("error in getting gcr credentials", "err", err)
			return "", err
		}
		dockerAuth.Username = gcrUsername
		dockerAuth.Password = gcrPassword
	default:
		// for all other registry types because they support direct username and password for authentication unlike ecr and gcr
	}
	encodedRegistryAuth, err := dockerAuth.GetEncodedRegistryAuth()
	if err != nil {
		logrus.Error("error in getting base64 encoded registry auth", "err", err)
		return "", err
	}
	return encodedRegistryAuth, nil
}

// GetImageDigestByImage fetches imageDigest from image using docker api client
func GetImageDigestByImage(ctx context.Context, image string, dockerAuth *bean.DockerAuthConfig) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Error("error in creating new docker client with options", "err", err)
		return "", err
	}
	var encodedRegistryAuth string
	if dockerAuth != nil && dockerAuth.IsRegistryPrivate {
		encodedRegistryAuth, err = getEncodedRegistryAuthForPrivateRegistry(dockerAuth)
		if err != nil {
			logrus.Error("error in getting encoded registry auth for private registry", "registryType", dockerAuth.RegistryType, "isRegistryPrivate", dockerAuth.IsRegistryPrivate, "err", err)
			return "", err
		}
	}
	inspectOutput, err := cli.DistributionInspect(ctx, image, encodedRegistryAuth)
	if err != nil {
		logrus.Error("error in getting digest", "image", image, "err", err)
		return "", err
	}
	return inspectOutput.Descriptor.Digest.String(), nil
}
