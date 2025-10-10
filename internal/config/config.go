package config

import (
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig
	Auth   AuthConfig
	DB     DBConfig
	CORS   CORSConfig
	Kuber  KuberConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type AuthConfig struct {
	KeycloakURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Realm        string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type KuberConfig struct {
	KubeConfigPath string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "3000"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Auth: AuthConfig{
			KeycloakURL:  getEnv("KEYCLOAK_URL", ""),
			ClientID:     getEnv("CLIENT_ID", ""),
			ClientSecret: getEnv("CLIENT_SECRET", ""),
			RedirectURL:  getEnv("REDIRECT_URL", ""),
			Realm:        getEnv("REALM", ""),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnv("DB_PORT", ""),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASS", ""),
			DBName:   getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "*"), ","),
			AllowedMethods:   strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
			AllowedHeaders:   strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Requested-With"), ","),
			AllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true",
		},
		Kuber: KuberConfig{
			KubeConfigPath: getEnv("KUBE_CONFIG_PATH", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ProvideConfig() *Config {
	cfg := Load()
	return &cfg
}
