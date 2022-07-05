package gostorage

import (
	"fmt"
	"os"
)

// GoStorageObject This type serves as an abstraction of a unit of storage (Local file, S3/Google Storage Object)
type GoStorageObject struct {
	Bucket        string
	Key           string
	Region        string
	IsLocal       bool
	LocalFilePath string
	ProviderType  ProviderType
}

type ProviderType string

const (
	ProviderAWS    ProviderType = "AWS"
	ProviderGoogle ProviderType = "Google"
)

func (receiver GoStorageObject) GetProvider(credentialsHolder CredentialsHolder) Provider {
	switch receiver.ProviderType {
	case ProviderGoogle:
		return GoogleStorage{CredentialsHolder: credentialsHolder}
	case ProviderAWS:
		return AWSStorage{CredentialsHolder: credentialsHolder}
	default:
		fmt.Fprintln(os.Stderr, "Error:", "unable to create the respective provider for provider type", receiver.ProviderType)
		os.Exit(1)
		return nil
	}
}
