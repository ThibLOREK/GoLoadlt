package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppName        string
	AppEnv         string
	HTTPPort       int
	PostgresDSN    string
	RedisAddr      string
	JWTSecret      string
	FrontendOrigin string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	viper.SetDefault("APP_NAME", "go-etl-studio")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("HTTP_PORT", 8080)
	viper.SetDefault("FRONTEND_ORIGIN", "http://localhost:3000")

	viper.AutomaticEnv()

	cfg := Config{
		AppName:        viper.GetString("APP_NAME"),
		AppEnv:         viper.GetString("APP_ENV"),
		HTTPPort:       viper.GetInt("HTTP_PORT"),
		PostgresDSN:    viper.GetString("POSTGRES_DSN"),
		RedisAddr:      viper.GetString("REDIS_ADDR"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		FrontendOrigin: viper.GetString("FRONTEND_ORIGIN"),
	}

	return cfg, nil
}
