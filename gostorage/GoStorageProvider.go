package gostorage

type Provider interface {
	createBucket(bucketName string, region string)
	deleteBucket(target GoStorageObject, deleteIfNotEmpty bool)

	copyFileWithinProvider(source GoStorageObject, target GoStorageObject)
	copyBucketWithinProvider(source GoStorageObject, target GoStorageObject)

	uploadFile(target GoStorageObject, sourceFile string)
	downloadFile(source GoStorageObject, targetFile string)
	listFilesInBucket(source GoStorageObject) []string

	deleteFile(target GoStorageObject)
}
