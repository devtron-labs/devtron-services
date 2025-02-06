package aws_v2

import (
	"os"
	"strconv"
)

const BucketName = "devtron-test"
const FileNameWithExtension = FileName + FileExtension
const FileName = "500mb-fake"
const FileExtension = ".pdf"

const OneMB = 1024 * 1024

const AwsSecretKey = "AWS_SECRET_KEY"
const AwsAccessKey = "AWS_ACCESS_KEY"

func GetPartSize() int64 {
	i, err := strconv.ParseInt(os.Getenv("PART_SIZE"), 10, 64)
	if err != nil {
		panic(err)
	}
	return OneMB * i
}

func GetConcurrency() int {
	i, err := strconv.Atoi(os.Getenv("CONCURRENCY"))
	if err != nil {
		panic(err)
	}
	return i
}
