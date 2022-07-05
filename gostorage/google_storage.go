package gostorage

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GoogleStorage struct {
	CredentialsHolder CredentialsHolder
}

var client *storage.Client

func (g GoogleStorage) createBucket(bucketName string, region string) {
	bucketHandle := g.getClient().Bucket(bucketName)
	_, err := bucketHandle.Attrs(context.Background())
	if err != nil && err == storage.ErrBucketNotExist {
		//shared.Log(shared.ProviderGoogle, fmt.Sprintf("Bucket %v doesn't exist, creating new one", shared.ArchiveBucketName))
		bucketLocationAttr := &storage.BucketAttrs{Location: DefaultGoogleRegion}
		if region != "" {
			bucketLocationAttr.Location = region
		}
		err = bucketHandle.Create(context.Background(), viper.GetString(GoogleProjectId), bucketLocationAttr)
		checkErr(err, fmt.Sprintf("unable to create bucket on GCP, Error %v", err))
	} else {
		checkErr(err, fmt.Sprintf("unable to access bucket on GCP, Error %v", err))
	}
}

func (g GoogleStorage) deleteBucket(target GoStorageObject, deleteIfNotEmpty bool) {
	filesInBucket := g.listFilesInBucket(target)
	storageClient := g.getClient()
	bucket := storageClient.Bucket(target.Bucket)

	if len(filesInBucket) > 0 && !deleteIfNotEmpty {
		fmt.Printf("bucket %v is not empty and should not be deleted, if you want to delete a bucket with its contents set deleteIfNotEmpty=true", target.Bucket)
		return
	}
	for _, key := range filesInBucket {
		g.deleteFile(GoStorageObject{Bucket: target.Bucket, Key: key, Region: target.Region})
	}
	err := bucket.Delete(context.Background())
	checkErr(err, fmt.Sprintf("unable to delete bucket on GCP, Error %v", err))
}

func (g GoogleStorage) uploadFile(target GoStorageObject, sourceFile string) {
	_, err := os.Stat(sourceFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "Error:", fmt.Sprintf("file %v does not exist", sourceFile))
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(sourceFile)
	checkErr(err, fmt.Sprintf("unable to read file content from %v, Error: %v", sourceFile, err))

	writer := g.getClient().Bucket(target.Bucket).Object(target.Key).NewWriter(context.Background())
	defer writer.Close()
	_, err = writer.Write(file)
	checkErr(err, fmt.Sprintf("unable to write to google storage, Error: %v", err))
}

func (g GoogleStorage) downloadFile(source GoStorageObject, targetFile string) {
	reader, err := g.getClient().Bucket(source.Bucket).Object(source.Key).NewReader(context.Background())
	checkErr(err, fmt.Sprintf("unable to read from google storage object %v, Error: %v", source.Key, err))
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", fmt.Sprintf("unable to download contents google storage object, Error: %v", err))
		os.Exit(1)
	} else {
		err = ioutil.WriteFile(targetFile, data, 0)
		checkErr(err, fmt.Sprintf("unable to write to file %v, Error: %v", targetFile, err))
	}
}

func (g GoogleStorage) listFilesInBucket(source GoStorageObject) []string {
	var keys []string
	objectIterator := g.getClient().Bucket(source.Bucket).Objects(context.Background(), nil)
	for {
		item, err := objectIterator.Next()
		if err == iterator.Done {
			break
		}
		checkErr(err, err)
		keys = append(keys, item.Name)
	}
	return keys
}

func (g GoogleStorage) deleteFile(target GoStorageObject) {
	err := g.getClient().Bucket(target.Bucket).Object(target.Key).Delete(context.Background())
	checkErr(err, fmt.Sprintf("unable to delete file %v, Error: %v", target.Key, err))
}

func (g GoogleStorage) copyFileWithinProvider(source GoStorageObject, target GoStorageObject) {
	src := g.getClient().Bucket(source.Bucket).Object(source.Key)
	dst := g.getClient().Bucket(target.Bucket).Object(target.Key)

	_, err := dst.CopierFrom(src).Run(context.Background())
	checkErr(err, fmt.Sprintf("unable to copy object from %v to %v, Error: %v", src, dst, err))
}

func (g GoogleStorage) copyBucketWithinProvider(source GoStorageObject, target GoStorageObject) {
	objectIterator := g.getClient().Bucket(source.Bucket).Objects(context.Background(), nil)
	for {
		item, err := objectIterator.Next()
		if err == iterator.Done {
			break
		}
		checkErr(err, err)
		source.Key = item.Name
		target.Key = item.Name
		g.copyFileWithinProvider(source, target)
	}
}

func (g GoogleStorage) getClient() *storage.Client {
	if client != nil {
		return client
	}
	client, err := storage.NewClient(context.Background(), option.WithCredentials(g.CredentialsHolder.GoogleCredentials))
	checkErr(err, fmt.Sprintf("unable to create Google storage client, Error: %v", err))
	return client
}
