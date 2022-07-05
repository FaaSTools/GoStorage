package gostorage

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
)

func checkErr(err interface{}, msg interface{}) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(1)
	}
}

func readFile(fileLocation string) []byte {
	file, err := ioutil.ReadFile(fileLocation)
	checkErr(err, fmt.Sprintf("unable to read file content from %v, Error: %v", fileLocation, err))
	return file
}
func getAbsolutePath(file *os.File) string {
	absolutePath, err := filepath.Abs(file.Name())
	checkErr(err, fmt.Sprintf("unable get absolute path from file %v, Error: %v", file.Name(), err))
	return absolutePath
}

func LoadCredentialsFromDefaultLocation() (*aws.Credentials, *google.Credentials) {
	var wd string

	/*
		Workaround for Google Cloud functions, as Google changes the structure of the deployed package it is not possible to assume that
		the credential files are in the current working directory. After some investigation I found out, that Google puts the code in a subdirectory
		/src/<packageName>
	*/
	googleCredentialsDirectory := "./src/p"
	info, err := os.Stat(googleCredentialsDirectory)
	googleFolderExists := !os.IsNotExist(err) && info.IsDir()

	if googleFolderExists {
		wd = googleCredentialsDirectory
	} else {
		wd, err = os.Getwd()
		checkErr(err, err)
	}

	//Set type for all configuration files to .yaml
	viper.AddConfigPath(wd)
	viper.SetConfigType("yaml")
	viper.SetConfigName("gcp-credentials")
	err = viper.MergeInConfig()
	checkErr(err, fmt.Sprintf("unable to find credentials file {%v}, Error: %v", "gcp-credentials", err))

	viper.SetConfigName("aws-credentials")
	err = viper.MergeInConfig()
	checkErr(err, fmt.Sprintf("unable to find credentials file {%v}, Error: %v", "aws-credentials", err))

	googleCredentials, err := google.CredentialsFromJSON(
		context.Background(),
		readFile(path.Join(wd, "gcp-credentials.yaml")),
		"https://www.googleapis.com/auth/devstorage.full_control",
	)
	checkErr(err, err)

	awsCredentials := &aws.Credentials{
		AccessKeyID:     viper.GetString(AWSAccessKey),
		SecretAccessKey: viper.GetString(AWSSecretAccessKey),
		SessionToken:    viper.GetString(AWSSessionTokenKey),
	}
	return awsCredentials, googleCredentials
}

// parseUrlToGoStorageObject parses Object/Bucket URLs from AWS and Google to extract information such as bucketName, key, region etc.
func parseUrlToGoStorageObject(urlString string) GoStorageObject {
	if isAWSUrl(urlString) {
		return parseAWSUrl(urlString)
	} else if isGoogleUrl(urlString) {
		return parseGoogleUrl(urlString)
	} else {
		if _, err := os.Stat(urlString); errors.Is(err, os.ErrNotExist) {
			checkErr(err, fmt.Sprintf("unable to find local file from {%v}, Error: %v", urlString, err))
		}
		return GoStorageObject{IsLocal: true, LocalFilePath: urlString}
	}
}

// parseAWSUrl AWS Object URL (with explicit region)
func parseAWSUrl(urlString string) GoStorageObject {
	var bucket string
	var key string
	var region string

	urlString = urlString[strings.Index(urlString, "https://")+len("https://"):]
	bucket = urlString[:strings.Index(urlString, ".")]
	urlString = urlString[strings.Index(urlString, ".")+len(".s3."):]
	if strings.HasPrefix(urlString, "amazonaws.com") { //No region specified
		region = DefaultAWSRegion
	} else {
		region = urlString[:strings.Index(urlString, ".")]
		urlString = urlString[strings.Index(urlString, ".")+1:]
	}
	urlString = urlString[strings.Index(urlString, "amazonaws.com")+len("amazonaws.com"):]
	if strings.HasPrefix(urlString, "/") {
		urlString = urlString[1:]
	}
	key = urlString
	return GoStorageObject{Bucket: bucket, Key: key, Region: region, ProviderType: ProviderAWS}
}

// parseGoogleUrl Google Object URL
func parseGoogleUrl(urlString string) GoStorageObject {
	var bucket string
	var key string

	if strings.HasPrefix(urlString, "gs://") {
		urlString = urlString[strings.Index(urlString, "gs://")+len("gs://"):]
	} else if strings.HasPrefix(urlString, "https://storage.cloud.google.com/") {
		urlString = urlString[strings.Index(urlString, "https://storage.cloud.google.com/")+len("https://storage.cloud.google.com/"):]
	}
	if strings.Contains(urlString, "/") {
		bucket = urlString[:strings.Index(urlString, "/")]
		key = urlString[strings.Index(urlString, "/")+1:]
	} else {
		bucket = urlString
	}
	return GoStorageObject{Bucket: bucket, Key: key, ProviderType: ProviderGoogle}
}

// isGoogleUrl Google Object URL: gs://gostorage-bucket-test/test.png
// Google Object URL: https://storage.cloud.google.com/gostorage-bucket-test/test.png
func isGoogleUrl(urlString string) bool {
	return strings.HasPrefix(urlString, "gs://") || strings.Contains(urlString, "storage.cloud.google.com")
}

// isAWSUrl AWS Object URL: https://gostorage-bucket-test.s3.amazonaws.com/newfile.png
func isAWSUrl(urlString string) bool {
	return strings.HasPrefix(urlString, "https://") && strings.Contains(urlString, "s3")
}
