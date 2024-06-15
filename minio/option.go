package minio

import (
	"net/http"
	"net/http/httptrace"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Option func(*minio.Options)

func WithSSL(useSSL bool) Option {
	return func(o *minio.Options) {
		o.Secure = useSSL
	}
}

func WithRegion(region string) Option {
	return func(o *minio.Options) {
		o.Region = region
	}
}

func WithCredentials(accessKey, secretKey string) Option {
	return func(o *minio.Options) {
		o.Creds = credentials.NewStaticV4(accessKey, secretKey, "")
	}
}

func WithTransport(transport *http.Transport) Option {
	return func(o *minio.Options) {
		o.Transport = transport
	}
}

func WithTrace(tracer *httptrace.ClientTrace) Option {
	return func(o *minio.Options) {
		o.Trace = tracer
	}
}
