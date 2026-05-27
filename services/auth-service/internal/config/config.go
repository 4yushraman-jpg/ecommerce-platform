package config

import "github.com/caarlos0/env/v11"

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`

	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`
	RequestTimeoutSeconds int `env:"REQUEST_TIMEOUT_SECONDS" envDefault:"15"`

	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`

	JWTSecret string `env:"JWT_SECRET,required"`
	JWTAccessTTLMinutes int `env:"JWT_ACCESS_TTL_MINUTES" envDefault:"15"`
	JWTRefreshTTLDays int `env:"JWT_REFRESH_TTL_DAYS" envDefault:"7"`

	CORSAllowedOrigin string `env:"CORS_ALLOWED_ORIGIN" envDefault:"*"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
