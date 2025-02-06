package main

import (
	"bytes"
	"context"
	"fmt"
	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	aws_v2 "github.com/devtron-labs/common-lib/blob-storage/aws-v2"
	"time"

	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"log"
	"os"
	"path/filepath"
)

const MaxUploadParts = 10000
const DefaultUploadConcurrency = 5
const DefaultUploadPartSize = MinUploadPartSize
const MinUploadPartSize int64 = 1024 * 1024 * 5

// above 50 its causing issues -https://github.com/aws/aws-sdk-go/issues/1763, maybe unrelated as its using waitgroup
const DefaultDownloadConcurrency = 5
const DefaultDownloadPartSize = 1024 * 1024 * 5

func UploadToS3Bucket() {

	filename := aws_v2.FileName + "-sdk-v1" + aws_v2.FileExtension
	bucket := aws_v2.BucketName
	item := aws_v2.FileName + "-sdk-v1-upload" + aws_v2.FileExtension
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("ap-south-1")})

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.PartSize = aws_v2.GetPartSize()
		d.Concurrency = aws_v2.GetConcurrency()
	})

	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println("error in reading source file", "sourceFile", filename, "destinationKey", item, "err", err)
		return
	}

	start := time.Now()

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(item),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("application/zip"),
	})
	if err != nil {
		fmt.Errorf("failed to upload file, %v", err)
	}
	elapsed := time.Since(start)
	log.Printf("upload took %s", elapsed)
	fmt.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
}

func DownloadFromS3Bucket() {

	bucket := aws_v2.BucketName
	item := aws_v2.FileNameWithExtension
	downloadFile := aws_v2.FileName + "-sdk-v1" + aws_v2.FileExtension

	file, err := os.Create(downloadFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{Region: aws.String("ap-south-1")})
	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = aws_v2.GetPartSize()
		d.Concurrency = aws_v2.GetConcurrency()
	})
	start := time.Now()
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		fmt.Println(err)
	}
	elapsed := time.Since(start)
	log.Printf("download-v1 took %s %s", elapsed, file.Name())

	fmt.Println("Downloaded-v1", file.Name(), numBytes, "bytes")
}

func main() {
	log.Println("config", "partSize", aws_v2.GetPartSize(), "concurrency", aws_v2.GetConcurrency())
	isUpload := os.Getenv("IS_UPLOAD")

	if isUpload == "true" {
		log.Println("Starting Uploading file from S3 bucket")
		uploadUsingDevtronCode()
		UploadToS3Bucket()
		aws_v2.RunGetStartedScenario(context.Background(), aws2.Config{Region: "ap-south-1"}, true)
	} else {
		log.Println("Starting  downloading file from S3 bucket")
		//downloadUsingDevtronCode()
		DownloadFromS3Bucket()
		aws_v2.RunGetStartedScenario(context.Background(), aws2.Config{Region: "ap-south-1"}, false)
	}

}

func downloadUsingDevtronCode() {

	//downloadSuccess := false
	//numBytes := int64(0)

	//s3: //aws-sam-cli-managed-default-samclisourcebucket-lwnny2w7mug8/cache-test/Argo CD 2.12.0.zip

	request := &blob_storage.BlobStorageRequest{
		StorageType:    blob_storage.BLOB_STORAGE_S3,
		SourceKey:      aws_v2.FileNameWithExtension,
		DestinationKey: aws_v2.FileName + "-devtron-download" + aws_v2.FileExtension,
		AwsS3BaseConfig: GetBlobStorageBaseS3Config(&blob_storage.BlobStorageS3Config{
			AccessKey:                  "",
			Passkey:                    "qWGO4K1kWYfmxqhZftRWWRsPcCcOQV2i6zRoRGmL",
			EndpointUrl:                "",
			IsInSecure:                 false,
			CiLogBucketName:            "devtron-test",
			CiLogRegion:                "ap-south-1",
			CiLogBucketVersioning:      false,
			CiCacheBucketName:          "devtron-test",
			CiCacheRegion:              "ap-south-1",
			CiCacheBucketVersioning:    false,
			CiArtifactBucketName:       "",
			CiArtifactRegion:           "",
			CiArtifactBucketVersioning: false,
		}, BlobStorageObjectTypeCache),
		AzureBlobBaseConfig: nil,
		GcpBlobBaseConfig:   nil,
	}

	file, err := os.Create(filepath.Clean(request.DestinationKey))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	//awsS3Blob := blob_storage.AwsS3Blob{}
	//downloadSuccess, numBytes, err = awsS3Blob.DownloadBlob(request, downloadSuccess, numBytes, err, file)
	blob_storage.DownLoadFromS3(file, request, nil)
}

func uploadUsingDevtronCode() {
	//s3: //aws-sam-cli-managed-default-samclisourcebucket-lwnny2w7mug8/cache-test/Argo CD 2.12.0.zip
	awsS3Blob := blob_storage.AwsS3Blob{}
	request := &blob_storage.BlobStorageRequest{
		StorageType:    blob_storage.BLOB_STORAGE_S3,
		SourceKey:      aws_v2.FileName + "-sdk-v1" + aws_v2.FileExtension,
		DestinationKey: aws_v2.FileName + "-devtron-upload" + aws_v2.FileExtension,
		AwsS3BaseConfig: GetBlobStorageBaseS3Config(&blob_storage.BlobStorageS3Config{
			AccessKey:                  os.Getenv(aws_v2.AwsAccessKey),
			Passkey:                    os.Getenv(aws_v2.AwsSecretKey),
			EndpointUrl:                "",
			IsInSecure:                 false,
			CiLogBucketName:            aws_v2.BucketName,
			CiLogRegion:                "ap-south-1",
			CiLogBucketVersioning:      false,
			CiCacheBucketName:          aws_v2.BucketName,
			CiCacheRegion:              "ap-south-1",
			CiCacheBucketVersioning:    false,
			CiArtifactBucketName:       "",
			CiArtifactRegion:           "",
			CiArtifactBucketVersioning: false,
		}, BlobStorageObjectTypeCache),
		AzureBlobBaseConfig: nil,
		GcpBlobBaseConfig:   nil,
	}

	file, err := os.Create(filepath.Clean(request.DestinationKey))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	//awsS3Blob := blob_storage.AwsS3Blob{}
	//downloadSuccess, numBytes, err = awsS3Blob.DownloadBlob(request, downloadSuccess, numBytes, err, file)
	err = awsS3Blob.UploadBlob(request, nil)
	if err != nil {
		log.Println(" -----> push err", err)
	}
}

const (
	BlobStorageObjectTypeCache    = "cache"
	BlobStorageObjectTypeArtifact = "artifact"
	BlobStorageObjectTypeLog      = "log"
)

func GetBlobStorageBaseS3Config(b *blob_storage.BlobStorageS3Config, blobStorageObjectType string) *blob_storage.AwsS3BaseConfig {
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:   os.Getenv(aws_v2.AwsAccessKey),
		Passkey:     os.Getenv(aws_v2.AwsSecretKey),
		EndpointUrl: b.EndpointUrl,
		IsInSecure:  b.IsInSecure,
	}
	switch blobStorageObjectType {
	case BlobStorageObjectTypeCache:
		awsS3BaseConfig.BucketName = b.CiCacheBucketName
		awsS3BaseConfig.Region = b.CiCacheRegion
		awsS3BaseConfig.VersioningEnabled = b.CiCacheBucketVersioning
		return awsS3BaseConfig
	case BlobStorageObjectTypeLog:
		awsS3BaseConfig.BucketName = b.CiLogBucketName
		awsS3BaseConfig.Region = b.CiLogRegion
		awsS3BaseConfig.VersioningEnabled = b.CiLogBucketVersioning
		return awsS3BaseConfig
	case BlobStorageObjectTypeArtifact:
		awsS3BaseConfig.BucketName = b.CiArtifactBucketName
		awsS3BaseConfig.Region = b.CiArtifactRegion
		awsS3BaseConfig.VersioningEnabled = b.CiArtifactBucketVersioning
		return awsS3BaseConfig
	default:
		return nil
	}
}
