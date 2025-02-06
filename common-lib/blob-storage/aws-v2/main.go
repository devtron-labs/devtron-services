package aws_v2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// RunGetStartedScenario is an interactive example that shows you how to use Amazon
// Simple Storage Service (Amazon S3) to create an S3 bucket and use it to store objects.
//
// 1. Create a bucket.
// 2. Upload a local file to the bucket.
// 3. Download an object to a local file.
// 4. Copy an object to a different folder in the bucket.
// 5. List objects in the bucket.
// 6. Delete all objects in the bucket.
// 7. Delete the bucket.
//
// This example creates an Amazon S3 service client from the specified sdkConfig so that
// you can replace it with a mocked or stubbed config for unit testing.
//
// It uses a questioner from the `demotools` package to get input during the example.
// This package can be found in the ..\..\demotools folder of this repo.
func RunGetStartedScenario(ctx context.Context, sdkConfig aws.Config, isUpload bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Something went wrong with the demo.")

		}
	}()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv(AwsAccessKey), os.Getenv(AwsSecretKey), "")))
	if err != nil {
		panic(err)
	}
	sdkConfig.Credentials = cfg.Credentials
	log.Println(strings.Repeat("-", 88))
	log.Println("Welcome to the Amazon S3 getting started demo.")
	log.Println(strings.Repeat("-", 88))

	s3Client := s3.NewFromConfig(sdkConfig)
	bucketBasics := BucketBasics{S3Client: s3Client}

	//count := 10
	//log.Printf("Let's list up to %v buckets for your account:", count)
	//buckets, err := bucketBasics.ListBuckets(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//if len(buckets) == 0 {
	//	log.Println("You don't have any buckets!")
	//} else {
	//	if count > len(buckets) {
	//		count = len(buckets)
	//	}
	//	for _, bucket := range buckets[:count] {
	//		log.Printf("\t%v\n", *bucket.Name)
	//	}
	//}

	bucketName := BucketName
	bucketExists, err := bucketBasics.BucketExists(ctx, bucketName)
	//if err != nil {
	//	panic(err)
	//}
	if !bucketExists {
		err = bucketBasics.CreateBucket(ctx, bucketName, sdkConfig.Region)
		if err != nil {
			panic(err)
		} else {
			log.Println("Bucket created.")
		}
	}
	log.Println(strings.Repeat("-", 88))
	smallKey := FileNameWithExtension
	downloadFileName := FileName + "-sdk-v1" + FileExtension
	uploadFileName := FileName + "-sdk-v2-upload" + FileExtension

	if isUpload {
		fmt.Println("Let's upload a file to your bucket.")
		smallFile := downloadFileName
		content, err := os.ReadFile(smallFile)
		if err != nil {
			log.Println("error in reading source file", "sourceFile", smallFile, "destinationKey", smallFile, "err", err)
			return
		}

		err = bucketBasics.UploadLargeObject(ctx, bucketName, uploadFileName, content)
		if err != nil {
			panic(err)
		}
		log.Printf("Uploaded %v as %v.\n", smallFile, uploadFileName)
		log.Println(strings.Repeat("-", 88))
	} else {
		log.Printf("Let's download %v to a file.", smallKey)
		bytes, err := bucketBasics.DownloadLargeObject(ctx, bucketName, smallKey)
		if err != nil {
			panic(err)
		}
		if bytes != nil {
			err = os.WriteFile(downloadFileName, bytes, 777)
			if err != nil {
				panic(err)
			}
		}

		log.Printf("File %v downloaded. -v2", downloadFileName)
		log.Println(strings.Repeat("-", 88))
	}

	//log.Printf("Let's copy %v to a folder in the same bucket.", smallKey)
	//folderName := questioner.Ask("Enter a folder name: ", demotools.NotEmpty{})
	//err = bucketBasics.CopyToFolder(ctx, bucketName, smallKey, folderName)
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("Copied %v to %v/%v.\n", smallKey, folderName, smallKey)
	//log.Println(strings.Repeat("-", 88))
	//
	//log.Println("Let's list the objects in your bucket.")
	//questioner.Ask("Press Enter when you're ready.")

	//objects, err := bucketBasics.ListObjects(ctx, bucketName)
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("Found %v objects.\n", len(objects))
	var objKeys []string
	//for _, object := range objects {
	//	objKeys = append(objKeys, *object.Key)
	//	log.Printf("\t%v\n", *object.Key)
	//}
	//log.Println(strings.Repeat("-", 88))

	if false {
		log.Println("Deleting objects.")
		err = bucketBasics.DeleteObjects(ctx, bucketName, objKeys)
		if err != nil {
			panic(err)
		}
		log.Println("Deleting bucket.")
		err = bucketBasics.DeleteBucket(ctx, bucketName)
		if err != nil {
			panic(err)
		}
		log.Printf("Deleting downloaded file %v.\n", downloadFileName)
		err = os.Remove(downloadFileName)
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("Okay. Don't forget to delete objects from your bucket to avoid charges.")
	}
	log.Println(strings.Repeat("-", 88))

	log.Println("Thanks for watching!")
	log.Println(strings.Repeat("-", 88))
}
