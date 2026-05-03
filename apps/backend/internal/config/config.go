// Package config liest und validiert die Backend-Config aus Defaults, optionaler
// YAML-Datei und ENV-Variablen (Prefix WWN_).
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config bündelt die gesamte Backend-Konfiguration.
type Config struct {
	HTTP           HTTPConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	Logging        LoggingConfig
	Environment    string
	MetricsEnabled bool
}

// HTTPConfig steuert den HTTP-Server.
type HTTPConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
}

// DatabaseConfig steuert die PostgreSQL-Pool-Verbindung.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig steuert den Redis-Client.
type RedisConfig struct {
	URL string
}

// LoggingConfig steuert das slog-Setup.
type LoggingConfig struct {
	Level  string
	Format string
}

// Load liest Defaults, optionale Config-Datei und ENV-Overrides ein und
// liefert eine validierte Config.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Leere Defaults für required-Keys, sonst greift AutomaticEnv() bei
	// Unmarshal nicht (bekannter Viper-Quirk: nested Keys ohne Default
	// werden vom Mapstructure-Decoder ignoriert).
	v.SetDefault("database.url", "")
	v.SetDefault("redis.url", "")

	v.SetDefault("http.port", 8080)
	v.SetDefault("http.readTimeout", "10s")
	v.SetDefault("http.writeTimeout", "30s")
	v.SetDefault("http.idleTimeout", "120s")
	v.SetDefault("http.shutdownTimeout", "15s")
	v.SetDefault("http.corsOrigins", []string{"http://app.localhost"})
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 5)
	v.SetDefault("database.connMaxLifetime", "1h")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("environment", "production")
	v.SetDefault("metricsEnabled", true)

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	v.SetEnvPrefix("WWN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("database.url (WWN_DATABASE_URL) is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("redis.url (WWN_REDIS_URL) is required")
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("http.port must be between 1 and 65535")
	}
	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("logging.level must be debug|info|warn|error")
	}
	switch c.Logging.Format {
	case "json", "text":
	default:
		return fmt.Errorf("logging.format must be json or text")
	}
	if c.Environment != "dev" && c.Environment != "staging" && c.Environment != "production" {
		return fmt.Errorf("environment must be dev|staging|production")
	}
	return nil
}
