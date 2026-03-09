package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config groups all application configuration sections
type Config struct {
	Env     string
	Server  ServerConfig
	DB      DBConfig
	Redis   RedisConfig
	Session SessionConfig
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port         string
	TrustedProxy string
}

// DBConfig contains database connection string
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig defines Redis cache setting
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	CacheTTL time.Duration
}

// SessionConfig controls session cookie behavior
type SessionConfig struct {
	Lifetime     time.Duration
	HttpOnly     bool
	SecureCookie bool
}

// Load reads configuration from .env
func Load() *Config {

	//Configure Viper to read .env
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	//Default configuration values
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_TRUSTED_PROXY", "127.0.0.1")

	viper.SetDefault("DB_DSN", "postgres://urluser:urlpassword@localhost:5432/urldb?sslmode=disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_CACHE_TTL", "24h")

	viper.SetDefault("SESSION_LIFETIME", "12h")
	viper.SetDefault("SESSION_HTTP_ONLY", true)
	viper.SetDefault("SESSION_SECURE_COOKIE", false)

	//Attempt to load .env file
	if err := viper.ReadInConfig(); err != nil {
		log.Println("no .env file found, using environment variables and defaults")
	}
	// Parse database connection lifetime
	connMaxLifetime, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		log.Fatalf("invalid DB_CONN_MAX_LIFETIME: %v", err)
	}

	// Parse Redis cache TTL
	cacheTTL, err := time.ParseDuration(viper.GetString("REDIS_CACHE_TTL"))
	if err != nil {
		log.Fatalf("invalid REDIS_CACHE_TTL: %v", err)
	}

	// Parse session lifetime
	sessionLifetime, err := time.ParseDuration(viper.GetString("SESSION_LIFETIME"))
	if err != nil {
		log.Fatalf("invalid SESSION_LIFETIME: %v", err)
	}

	// Construct and return application config
	return &Config{
		Env: strings.ToLower(viper.GetString("ENV")),

		Server: ServerConfig{
			// Ensure port always has ":" prefix
			Port:         ":" + strings.TrimPrefix(viper.GetString("SERVER_PORT"), ":"),
			TrustedProxy: viper.GetString("SERVER_TRUSTED_PROXY"),
		},

		DB: DBConfig{
			DSN:             viper.GetString("DB_DSN"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: connMaxLifetime,
		},

		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
			CacheTTL: cacheTTL,
		},

		Session: SessionConfig{
			Lifetime:     sessionLifetime,
			HttpOnly:     viper.GetBool("SESSION_HTTP_ONLY"),
			SecureCookie: viper.GetBool("SESSION_SECURE_COOKIE"),
		},
	}
}
