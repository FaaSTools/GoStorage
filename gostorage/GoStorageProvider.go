package gostorage

import "io"

type Provider interface {
	createBucket(bucketName string, region string)
	deleteBucket(target GoStorageObject, deleteIfNotEmpty bool)

	copyFileWithinProvider(source GoStorageObject, target GoStorageObject)
	copyBucketWithinProvider(source GoStorageObject, target GoStorageObject)

	uploadFile(target GoStorageObject, sourceFile string)
	downloadFile(source GoStorageObject, targetFile string)
	downloadFileAsReader(source GoStorageObject) io.Reader
	listFilesInBucket(source GoStorageObject) []string

	deleteFile(target GoStorageObject)
}
