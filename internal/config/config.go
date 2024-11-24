package config

import (
	"context"
	"path"
	"time"

	"github.com/catalystgo/logger/logger"
	"github.com/spf13/viper"
)

type AppConfig struct {
	App struct {
		Name string `yaml:"name"`
	} `yaml:"app"`
	Server struct {
		Admin struct {
			Port int `yaml:"port"`
		} `yaml:"admin"`
		HTTP struct {
			Port int `yaml:"port"`
		} `yaml:"http"`
		Grpc struct {
			Port int `yaml:"port"`
		} `yaml:"grpc"`
		Shutdown struct {
			Timeout time.Duration `yaml:"timeout"`
			Delay   time.Duration `yaml:"delay"`
		} `yaml:"shutdown"`
	} `yaml:"server"`
}

func Parse(ctx context.Context, file string) (*AppConfig, error) {
	var cfg AppConfig

	viper.SetConfigName(path.Base(file))
	viper.AddConfigPath(path.Dir(file))
	viper.SetConfigType("yaml")

	logger.Infof(ctx, "load config from path: %s", file)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
