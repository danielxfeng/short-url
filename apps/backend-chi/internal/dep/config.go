package dep

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type AppModeType string

const (
	EnvProd AppModeType = "production"
	EnvDev  AppModeType = "development"
	EnvTest AppModeType = "test"
)

var allowedAppModes = []AppModeType{EnvProd, EnvDev, EnvTest}

type Config struct {
	AppMode      AppModeType
	Port         int
	Cors         string
	SentryDSN    string
	DbURL        string
	TestDbURL    string
	JWTSecret    string
	JWTExpiry    time.Duration
	NotFoundPage string
}

func GetEnvStrOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func GetEnvStrOrError(key string) (string, error) {
	value := os.Getenv(key)

	if value == "" {
		return "", fmt.Errorf("environment variable %s is required but not set", key)
	}

	return value, nil
}

func GetEnvIntOrDefault(key string, defaultValue int) int {
	strValue := os.Getenv(key)

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func LoadAppMode() AppModeType {
	appModeStr := GetEnvStrOrDefault("APP_MODE", "development")

	appModeStr = strings.ToLower(strings.TrimSpace(appModeStr))

	if !slices.Contains(allowedAppModes, AppModeType(appModeStr)) {
		return EnvDev
	}

	return AppModeType(appModeStr)
}

func LoadConfigFromEnv() (*Config, error) {
	jwtSecret, err := GetEnvStrOrError("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppMode:      LoadAppMode(),
		Port:         GetEnvIntOrDefault("PORT", 8080),
		Cors:         GetEnvStrOrDefault("CORS", "http://localhost:5173"),
		SentryDSN:    GetEnvStrOrDefault("SENTRY_DSN", ""),
		DbURL:        GetEnvStrOrDefault("DB_URL", "postgresql://user:password@localhost:5432/dbname?sslmode=disable"),
		TestDbURL:    GetEnvStrOrDefault("TEST_DB_URL", "postgresql://user:password@localhost:5432/test_dbname?sslmode=disable"),
		JWTSecret:    jwtSecret,
		JWTExpiry:    time.Duration(GetEnvIntOrDefault("JWT_EXPIRY", 24*7)) * time.Hour, // default to 7 days
		NotFoundPage: GetEnvStrOrDefault("NOT_FOUND_PAGE", "http://localhost:5173/not-found"),
	}

	return cfg, nil
}
