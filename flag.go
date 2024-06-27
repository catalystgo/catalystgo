package catalystgo

import (
	"context"
	"flag"
	"os"

	"github.com/catalystgo/catalystgo/errors"
	"github.com/catalystgo/logger/logger"
)

var (
	configPath = flag.String("config", "./catalystgo/config-development.yml", "CatalystGo config file")
)

func init() {
	flag.Parse()

	ctx := context.Background()

	if configPath == nil {
		logger.Fatalf(ctx, "config path not passed. consider using --config")
	}

	if _, err := os.Stat(*configPath); errors.Is(err, os.ErrNotExist) {
		logger.Fatalf(ctx, "config path not passed. consider using --config")
	}
}
