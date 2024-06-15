package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Create a new miniio client
func NewClient(endpoint, accessKey, secretKey string, opts ...Option) (*minio.Client, error) {
	// Set default oo
	oo := &minio.Options{
		Secure: true,
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
	}

	// Apply options
	for _, opt := range opts {
		opt(oo)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, oo)
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
