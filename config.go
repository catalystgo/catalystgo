package catalystgo

import (
	"path"
	"time"

	"github.com/spf13/viper"
)

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
			Delay   time.Duration `yaml:"delay"` // TODO: Use delay to delay the shutdown
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
func Parse() (*config, error) {
	c := &config{}
	if err := parseConfig(c); err != nil {
		return nil, err
	}
	return c, nil
}

// parseConfig parses the configuration file
// and unmarshals it into the config struct
func parseConfig(c *config) error {
	viper.SetConfigName(path.Base(*configPath))
	viper.AddConfigPath(path.Dir(*configPath))
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(c); err != nil {
		return err
	}

	return nil
}
