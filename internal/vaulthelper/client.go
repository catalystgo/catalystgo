package vaulthelper

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// env related constants
const (
	EnvVaultAddress = "VAULT_ADDR"
	EnvVaultAuthURL = "VAULT_AUTH_URL"
	EnvVaultToken   = "VAULT_TOKEN"
	EnvVaultRole    = "VAULT_ROLE"
)

// auth related constants
const (
	authPathPrefix = "auth/"
	authPathSuffix = "/login"

	defaultKubernetesAuthURL = "auth/kubernetes/login"
	serviceAccountTokenPath  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// timing and retry configuration
const (
	clientRequestTimeout = 90 * time.Second
	maxRetries           = 100
	minRetryInterval     = 1 * time.Second
	maxRetryInterval     = 8 * time.Second
	maxJitter            = 1000
)

const tokenRefreshInterval = 1 * time.Minute

type Config struct {
	Namespace string
	// JWTRole by default VAULT_ROLE or Namespace if VAULT_ROLE is empty.
	AuthURL string
	JWTRole string
	JWTPath string
	Vault   *api.Config
}

func NewConfig(namespace, authURL, jwtPath string) *Config {
	return &Config{
		Namespace: namespace,
		AuthURL:   buildAuthURL(authURL),
		JWTPath:   jwtPath,
		Vault:     api.DefaultConfig(),
	}
}

func NewDefaultConfig(namespace string) *Config {
	return NewConfig(namespace, determineAuthURL(), serviceAccountTokenPath)
}

func NewClient(config *Config, options ...Option) (*api.Client, error) {
	opts := newDefaultOptions()
	for _, o := range options {
		o(opts)
	}

	logger := opts.logger

	if err := validateConfig(config, logger); err != nil {
		return nil, err
	}

	logger.Debug(
		"vault_helper: vault config",
		zap.String("namespace", config.Namespace),
		zap.String("auth_url", config.AuthURL),
		zap.String("jwt_path", config.JWTPath),
		zap.String("vault_role", config.JWTRole),
		zap.String("vault_addr_env", os.Getenv(EnvVaultAddress)),
	)

	client, err := api.NewClient(config.Vault)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %v", err)
	}

	configureClientPolicy(client, logger)

	var secret *api.Secret
	if client.Token() == "" {
		secret, err = authenticate(client, config, logger)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		for {
			timeout := calculateUpdateTokenInterval(secret)
			time.Sleep(timeout)

			secret, err = authenticate(client, config, logger)
			if err != nil {
				continue
			}

			if secret == nil || secret.Auth == nil {
				continue
			}

			logger.Debug("vault_helper: token renewed", zap.String("token", hideToken(secret.Auth.ClientToken)))
		}
	}()

	return client, nil
}

func validateConfig(config *Config, logger *zap.Logger) error {
	if config == nil {
		logger.Error("vault_helper: config is nil")
		return errors.New("config is required")
	}

	if config.Namespace == "" {
		logger.Error("vault_helper: namespace is empty")
		return errors.New("namespace is required")
	}

	if config.AuthURL == "" {
		logger.Error("vault_helper: auth URL is empty")
		return errors.New("auth URL is required")
	}

	if config.JWTPath == "" {
		logger.Error("vault_helper: JWT path is empty")
		return errors.New("JWT path is required")
	}

	if config.JWTRole == "" {
		config.JWTRole = determineRole(config.Namespace)
	}

	return nil
}

func authenticate(client *api.Client, config *Config, logger *zap.Logger) (*api.Secret, error) {
	jwtToken, err := readJWTToken(config.JWTPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT token: %v", err)
	}

	authData := map[string]interface{}{
		"role": config.JWTRole,
		"jwt":  jwtToken,
	}

	tokenResp, err := client.Logical().Write(config.AuthURL, authData)
	if err != nil {
		return nil, fmt.Errorf("failed to request auth token: %v", err)
	}
	if tokenResp == nil || tokenResp.Auth == nil {
		return nil, errors.New("failed to request auth token: response is nil")
	}

	client.SetToken(tokenResp.Auth.ClientToken)
	logger.Debug("token obtained", zap.String("token", hideToken(tokenResp.Auth.ClientToken)))

	return tokenResp, nil
}

func configureClientPolicy(client *api.Client, logger *zap.Logger) {
	client.SetClientTimeout(clientRequestTimeout)
	client.SetMaxRetries(maxRetries)
	client.SetMinRetryWait(minRetryInterval)
	client.SetMaxRetryWait(maxRetryInterval)

	client.SetCheckRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			logger.Info("vault_helper: run backoff retry", zap.Int("statusCode", resp.StatusCode))
			return true, nil
		}
		return api.DefaultRetryPolicy(ctx, resp, err)
	})

	client.SetBackoff(calculateBackoff)
}

func calculateBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// берем за интервал степень двойки
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)

	// но время не должно быть больше чем max
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}

	jitter := int64(maxJitter)
	if n, err := rand.Int(rand.Reader, big.NewInt(1000)); err == nil {
		jitter = n.Int64()
	}

	jitterDuration := time.Duration(float64(jitter) * float64(time.Millisecond))
	return sleep + jitterDuration
}

func determineAuthURL() string {
	authURL := os.Getenv(EnvVaultAuthURL)
	if authURL == "" {
		return defaultKubernetesAuthURL
	}
	return buildAuthURL(authURL)
}

func buildAuthURL(authURL string) string {
	if !strings.HasPrefix(authURL, authPathPrefix) {
		authURL = authPathPrefix + authURL
	}
	if !strings.HasSuffix(authURL, authPathSuffix) {
		authURL += authPathSuffix
	}
	return authURL
}

func determineRole(namespace string) string {
	role := os.Getenv(EnvVaultRole)
	if role == "" {
		return namespace
	}
	return role
}

func readJWTToken(path string) (string, error) {
	token, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}

	if len(token) == 0 {
		return "", errors.New("JWT token is empty")
	}

	return string(token), nil
}

func hideToken(token string) string {
	if len(token) > 4 {
		return token[:2] + strings.Repeat("*", len(token)-4) + token[len(token)-2:]
	}
	return "****"
}

func calculateUpdateTokenInterval(secret *api.Secret) time.Duration {
	if secret == nil {
		return tokenRefreshInterval
	}
	ttl, err := secret.TokenTTL()
	if err != nil {
		return tokenRefreshInterval
	}
	return ttl / 2
}
