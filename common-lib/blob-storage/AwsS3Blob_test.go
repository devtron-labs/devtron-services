package blob_storage

import (
	"context"
	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"net/url"
	"os"
	"testing"
)

func TestBucketBasics_DownloadFileV2(t *testing.T) {
	type Fields struct {
		S3Client *s3v2.Client
	}
	type Args struct {
		ctx             context.Context
		request         *BlobStorageRequest
		downloadSuccess bool
		numBytes        int64
		err             error
		file            *os.File
	}
	//cfg, _ := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentialsv2.NewStaticCredentialsProvider("", "", "")))
	//region := "ap-south-1"
	//sdkConfig := awsv2.Config{Region: region}
	//sdkConfig.Credentials = cfg.Credentials
	//
	//s3Client := s3v2.NewFromConfig(sdkConfig)

	endpointURL, _ := url.Parse("http://34.47.202.209:9000") // or where ever you ran minio

	s3Client := s3v2.New(s3v2.Options{

		EndpointResolverV2: &Resolver{URL: endpointURL},
		Credentials: awsv2.CredentialsProviderFunc(func(ctx context.Context) (awsv2.Credentials, error) {
			return awsv2.Credentials{
				AccessKeyID:     "",
				SecretAccessKey: "",
			}, nil
		}),
	})
	tests := []struct {
		name    string
		fields  Fields
		args    Args
		want    bool
		want1   int64
		wantErr bool
	}{{name: "test-1", fields: Fields{S3Client: s3Client}, args: Args{
		ctx: context.Background(),
		request: &BlobStorageRequest{
			StorageType:    BLOB_STORAGE_S3,
			SourceKey:      "qa-devtroncd-8/120-ci-7-g5l2-9/main.log",
			DestinationKey: "job-artifact.zip",
			AwsS3BaseConfig: &AwsS3BaseConfig{
				AccessKey:   "",
				Passkey:     "",
				EndpointUrl: "http://34.47.202.209:9000",
				BucketName:  "test-bucket",
				Region:      "ap-south-1",
			},
		},
		downloadSuccess: false,
		numBytes:        0,
		err:             nil,
	}, want: true, want1: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basics := BucketBasics{
				S3Client: tt.fields.S3Client,
			}
			got, got1, err := basics.DownloadFileV2(tt.args.ctx, tt.args.request, tt.args.downloadSuccess, tt.args.numBytes, tt.args.err, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFileV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DownloadFileV2() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DownloadFileV2() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
