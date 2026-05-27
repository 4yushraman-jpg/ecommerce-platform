package config

import "github.com/caarlos0/env/v11"

type Config struct {
	AppEnv                string `env:"APP_ENV" envDefault:"development"`
	HTTPPort              string `env:"HTTP_PORT" envDefault:"8081"`
	RequestTimeoutSeconds int    `env:"REQUEST_TIMEOUT_SECONDS" envDefault:"15"`
	CORSAllowedOrigin     string `env:"CORS_ALLOWED_ORIGIN" envDefault:"*"`

	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`

	RedisAddr       string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword   string `env:"REDIS_PASSWORD"`
	RedisDB         int    `env:"REDIS_DB" envDefault:"0"`
	CacheTTLSeconds int    `env:"CACHE_TTL_SECONDS" envDefault:"120"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
