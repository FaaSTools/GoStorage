package gostorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	aws_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	types2 "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWSStorage struct {
	CredentialsHolder CredentialsHolder
}

func (a AWSStorage) createBucket(bucketName string, region string) {
	bucketInput := &aws_s3.CreateBucketInput{Bucket: &bucketName}
	if region != "" && region != DefaultAWSRegion {
		bucketInput.CreateBucketConfiguration = &types2.CreateBucketConfiguration{LocationConstraint: types2.BucketLocationConstraint(region)}
	}
	_, err := a.getClientWithRegion(region).CreateBucket(context.Background(), bucketInput)
	checkErr(err, fmt.Sprintf("unable to create bucket on AWS, Error %v", err))
}

func (a AWSStorage) deleteBucket(target GoStorageObject, deleteIfNotEmpty bool) {
	filesInBucket := a.listFilesInBucket(target)
	if len(filesInBucket) > 0 && !deleteIfNotEmpty {
		fmt.Printf("bucket %v is not empty and should not be deleted, if you want to delete a bucket with its contents set deleteIfNotEmpty=true", target.Bucket)
		return
	}

	for _, key := range filesInBucket {
		a.deleteFile(GoStorageObject{Bucket: target.Bucket, Key: key, Region: target.Region})
	}
	_, err := a.getClientWithRegion(target.Region).DeleteBucket(context.Background(), &aws_s3.DeleteBucketInput{Bucket: &target.Bucket})
	checkErr(err, fmt.Sprintf("unable to delete bucket on AWS, Error %v", err))
}

func (a AWSStorage) uploadFile(target GoStorageObject, sourceFile string) {
	_, err := os.Stat(sourceFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "Error:", fmt.Sprintf("file %v does not exist", sourceFile))
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(sourceFile)
	checkErr(err, fmt.Sprintf("unable to read file content from %v, Error: %v", sourceFile, err))

	_, err = a.getClientWithRegion(target.Region).PutObject(context.Background(), &aws_s3.PutObjectInput{Bucket: &target.Bucket, Key: &target.Key, Body: bytes.NewReader(file)})
	checkErr(err, fmt.Sprintf("unable to write to AWS, Error: %v", err))
}

func (a AWSStorage) downloadFile(source GoStorageObject, targetFile string) {
	getObjectOutput, err := a.getClientWithRegion(source.Region).GetObject(context.Background(), &aws_s3.GetObjectInput{Bucket: &source.Bucket, Key: &source.Key})
	checkErr(err, fmt.Sprintf("unable to read from AWS storage object %v, Error: %v", source.Key, err))
	defer getObjectOutput.Body.Close()

	data, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", fmt.Sprintf("unable to download contents AWS storage object, Error: %v", err))
		os.Exit(1)
	} else {
		err = ioutil.WriteFile(targetFile, data, 0)
		checkErr(err, fmt.Sprintf("unable to write to file %v, Error: %v", targetFile, err))
	}
}

func (a AWSStorage) downloadFileAsReader(source GoStorageObject) io.Reader {
	getObjectOutput, err := a.getClientWithRegion(source.Region).GetObject(context.Background(), &aws_s3.GetObjectInput{Bucket: &source.Bucket, Key: &source.Key})
	checkErr(err, fmt.Sprintf("unable to read from AWS storage object %v, Error: %v", source.Key, err))
	return getObjectOutput.Body
}

func (a AWSStorage) listFilesInBucket(source GoStorageObject) []string {
	var keys []string
	listObjects, err := a.getClientWithRegion(source.Region).ListObjectsV2(context.Background(), &aws_s3.ListObjectsV2Input{Bucket: &source.Bucket})
	checkErr(err, fmt.Sprintf("unable to list files from bucket %v, Error: %v", source.Bucket, err))
	for _, k := range listObjects.Contents {
		keys = append(keys, *k.Key)
	}
	return keys
}

func (a AWSStorage) deleteFile(target GoStorageObject) {
	_, err := a.getClientWithRegion(target.Region).DeleteObject(context.Background(), &aws_s3.DeleteObjectInput{Bucket: &target.Bucket, Key: &target.Key})
	checkErr(err, fmt.Sprintf("unable to delete file %v, Error: %v", target.Key, err))
}

func (a AWSStorage) copyFileWithinProvider(source GoStorageObject, target GoStorageObject) {
	storageClient := a.getClientWithRegion(source.Region)
	if target.Region != "" && target.Region != DefaultAWSRegion {
		storageClient = a.getClientWithRegion(target.Region)
	}
	sourceString := fmt.Sprintf("%v/%v", source.Bucket, source.Key)
	_, err := storageClient.CopyObject(context.Background(), &aws_s3.CopyObjectInput{Bucket: &target.Bucket, CopySource: &sourceString, Key: &target.Key})
	checkErr(err, fmt.Sprintf("unable to copy object from %v to %v, Error: %v", source.Bucket, target.Bucket, err))
}

func (a AWSStorage) copyBucketWithinProvider(source GoStorageObject, target GoStorageObject) {
	filesInBucket := a.listFilesInBucket(source)
	for _, fileKey := range filesInBucket {
		source.Key = fileKey
		target.Key = fileKey
		a.copyFileWithinProvider(source, target)
	}
}

func (a AWSStorage) getClientWithRegion(region string) *aws_s3.Client {
	if region == "" {
		region = DefaultAWSRegion
	}

	staticCredentialsProvider := credentials.StaticCredentialsProvider{Value: *a.CredentialsHolder.AwsCredentials}
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region), config.WithCredentialsProvider(staticCredentialsProvider))
	checkErr(err, fmt.Sprintf("unable to load AWS SDK config, Error: %v", err))
	return aws_s3.NewFromConfig(cfg)
}
