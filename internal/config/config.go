// backend/internal/config/config.go

package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	SecretKey        string        `mapstructure:"secret-key"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
	PasswordResetTTL time.Duration `mapstructure:"password_reset_ttl"`
	Issuer           string        `mapstructure:"issuer"`
}

// EmailConfig contém configurações para envio de email
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUsername string `mapstructure:"smtp_username"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	FromName     string `mapstructure:"from_name"`
	ResetURL     string `mapstructure:"reset_url"`
}

// LogConfig contém configurações de logging
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Email    EmailConfig    `mapstructure:"email"`
	Log      LogConfig      `mapstructure:"log"`
}

var cfg *Config

// defaultConfigPath is used when LoadConfig is called with no path.
const defaultConfigPath = "./configs"

// LoadConfig loads config from the default path (./configs). For tests or custom paths, use LoadConfigFromPath.
func LoadConfig() (*Config, error) {
	return LoadConfigFromPath(defaultConfigPath)
}

// LoadConfigFromPath loads config from the given directory (must contain app.yml).
// Pass "" to use defaultConfigPath. Used by tests with a temp dir to avoid touching ./configs.
// Resets viper state so only the given path is used (no leftover paths from previous loads).
func LoadConfigFromPath(configPath string) (*Config, error) {
	viper.Reset()
	if configPath == "" {
		configPath = defaultConfigPath
	}

	viper.SetConfigName("app")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("falha ao ler o arquivo de configuração: %w", err)
	}

	// DATABASE_DSN env overrides config file when set
	viper.AutomaticEnv()
	_ = viper.BindEnv("database.dsn", "DATABASE_DSN")

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("falha ao carregar as configurações: %w", err)
	}

	return cfg, nil
}

func GetConfig() *Config {
	return cfg
}
