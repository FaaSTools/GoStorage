package gostorage

import (
	"fmt"
	"os"
)

type GoStorage struct {
	Credentials CredentialsHolder
}

func (s GoStorage) CreateBucket(storageObject GoStorageObject) {
	storageObject.GetProvider(s.Credentials).createBucket(storageObject.Bucket, storageObject.Region)
}

func (s GoStorage) DeleteBucket(storageObject GoStorageObject, deleteIfNotEmpty bool) {
	storageObject.GetProvider(s.Credentials).deleteBucket(storageObject, deleteIfNotEmpty)
}

func (s GoStorage) CopyFromString(source string, target string) {
	sourceObject := parseUrlToGoStorageObject(source)
	targetObject := parseUrlToGoStorageObject(target)

	s.Copy(sourceObject, targetObject)
}

func (s GoStorage) ListFilesInBucketFromString(target string) []string {
	targetObject := parseUrlToGoStorageObject(target)

	return s.ListFilesInBucket(targetObject)
}

func (s GoStorage) Copy(source GoStorageObject, target GoStorageObject) {
	if source.IsLocal && !target.IsLocal { //Upload file
		target.GetProvider(s.Credentials).createBucket(target.Bucket, target.Region)
		target.GetProvider(s.Credentials).uploadFile(target, source.LocalFilePath)

	} else if !source.IsLocal && target.IsLocal { //Download file
		source.GetProvider(s.Credentials).downloadFile(source, target.LocalFilePath)

	} else if !source.IsLocal && !target.IsLocal { //Copy between (possibly different) providers
		target.GetProvider(s.Credentials).createBucket(target.Bucket, target.Region)

		if source.ProviderType == target.ProviderType {
			if source.Key == "" && target.Key == "" {
				target.GetProvider(s.Credentials).copyBucketWithinProvider(source, target)
			} else if source.Bucket != "" && source.Key != "" {
				target.GetProvider(s.Credentials).copyFileWithinProvider(source, target)
			}

		} else if source.ProviderType != target.ProviderType {
			if source.Key == "" && target.Key == "" {
				s.copyBucket(source, target)
			} else if source.Key != "" && target.Key != "" {
				s.copyFile(source, target)
			} else {
				fmt.Fprintln(os.Stderr, "Error:", "Incorrect configuration of source and target key found")
				os.Exit(1)
			}

		} else {
			fmt.Fprintln(os.Stderr, "Error:", "Incorrect configuration of source and target location found")
			os.Exit(1)
		}
	}
}

func (s GoStorage) ListFilesInBucket(target GoStorageObject) []string {
	return target.GetProvider(s.Credentials).listFilesInBucket(target)
}

func (s GoStorage) DeleteFile(target GoStorageObject) {
	target.GetProvider(s.Credentials).deleteFile(target)
}

func (s GoStorage) DeleteFileFromString(url string) {
	storageObject := parseUrlToGoStorageObject(url)
	storageObject.GetProvider(s.Credentials).deleteFile(storageObject)
}

func (s GoStorage) UploadFile(source GoStorageObject) {
	source.GetProvider(s.Credentials).uploadFile(source, source.LocalFilePath)
}

func (s GoStorage) DownloadFile(source GoStorageObject, targetFile string) {
	source.GetProvider(s.Credentials).downloadFile(source, targetFile)
}

// ---- Helper functions ----

func (s GoStorage) copyBucket(source GoStorageObject, target GoStorageObject) {
	filesInBucket := source.GetProvider(s.Credentials).listFilesInBucket(source)
	for _, file := range filesInBucket {
		source.Key = file
		target.Key = file
		s.copyFile(source, target)
	}
}

func (s GoStorage) copyFile(source GoStorageObject, target GoStorageObject) {
	tempFile, err := os.CreateTemp(os.TempDir(), source.Key)
	checkErr(err, fmt.Sprintf("unable to create temporary file, Error: %v", err))
	tempFilePath := getAbsolutePath(tempFile)

	source.GetProvider(s.Credentials).downloadFile(source, tempFilePath)

	target.GetProvider(s.Credentials).uploadFile(target, tempFilePath)

	err = os.Remove(tempFilePath)
	checkErr(err, fmt.Sprintf("unable to delete temp. file %v, Error: %v\n", tempFile.Name(), err))
}
