package catalystgo

import (
	"context"
	"errors"
	"flag"
	"os"
	"path"
	"time"

	"github.com/catalystgo/logger/logger"
	"github.com/spf13/viper"
)

var (
	configPath = flag.String("config", "./.catalystgo/config-development.yml", "CatalystGo config file")
)

func init() {
	flag.Parse()

	ctx := context.Background()

	if configPath == nil {
		logger.Fatal(ctx, "config path not passed. consider using --config")
	}

	if _, err := os.Stat(*configPath); errors.Is(err, os.ErrNotExist) {
		logger.Fatalf(ctx, "config path not valid: %v", err)
	}
}

type config struct {
	App struct {
		Name string `yaml:"name"`
	} `yaml:"app"`
	Server struct {
		Debug struct {
			Port int `yaml:"port"`
		} `yaml:"debug"`
		HTTP struct {
			Port int `yaml:"port"`
		} `yaml:"http"`
		Grpc struct {
			Port int `yaml:"port"`
		} `yaml:"grpc"`
		GracefulShutdown struct {
			Timeout time.Duration `yaml:"timeout"`
			Delay   time.Duration `yaml:"delay"`
		} `yaml:"graceful_shutdown"`
	} `yaml:"server"`
	Tracing struct {
		Enabled  bool   `yaml:"enabled"`
		Provider string `yaml:"provider"`
		Address  string `yaml:"address"`
	} `yaml:"tracing"`
	Vault struct {
		Enable  bool   `yaml:"enable"`
		Address string `yaml:"address"`
		Token   string `yaml:"token"`
	} `yaml:"vault"`
	RateLimiter struct {
		Enable  bool `yaml:"enable"`
		Default struct {
			Limit int `yaml:"limit"`
			Burst int `yaml:"burst"`
		} `yaml:"default"`
		Handlers []struct {
			Method string `yaml:"method"`
			Limit  int    `yaml:"limit"`
			Burst  int    `yaml:"burst"`
		} `yaml:"handlers"`
	} `yaml:"rate_limiter"`
	RealtimeConfig []struct {
		Name  string `yaml:"name"`
		Usage string `yaml:"usage"`
		Value string `yaml:"value"`
		Type  string `yaml:"type"`
	} `yaml:"realtime_config"`
	Env     map[string]string `yaml:"env"`
	Secrets map[string]string `yaml:"secrets"`
}

// Parse parses the framework configuration file
func Parse(ctx context.Context) (*config, error) {
	c := &config{}
	if err := parseConfig(ctx, c); err != nil {
		return nil, err
	}

	logger.Error(ctx, "cfg.vault.enable: %b", c.Vault.Enable)
	logger.Error(ctx, "cfg.tracing.enabled: %b", c.Tracing.Enabled)
	logger.Error(ctx, "cfg.rate_limiter.enable: %b", c.RateLimiter.Enable)

	return c, nil
}

// parseConfig parses the configuration file
// and unmarshals it into the config struct
func parseConfig(ctx context.Context, c *config) error {
	viper.SetConfigName(path.Base(*configPath))
	viper.AddConfigPath(path.Dir(*configPath))
	viper.SetConfigType("yaml")

	logger.Debugf(ctx, "loading config from path: %s", *configPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(c); err != nil {
		return err
	}

	return nil
}
