package vaulthelper

import (
	"go.uber.org/zap"
)

type helperOptions struct {
	logger *zap.Logger
}

func newDefaultOptions() *helperOptions {
	return &helperOptions{logger: zap.NewNop()}
}

type Option func(o *helperOptions)

func WithLogger(l *zap.Logger) Option {
	return func(o *helperOptions) {
		o.logger = l
	}
}
