package configs

import (
	"errors"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv         string        `env:"APP_ENV" envDefault:"development"`
	Port           int           `env:"APP_PORT" envDefault:"8080"`
	DatabaseURL    string        `env:"DATABASE_URL,required"`
	JWTSecret      string        `env:"JWT_SECRET,required"`
	JWTTTL         time.Duration `env:"JWT_TTL" envDefault:"8h"`
	HistoryDays    int           `env:"HISTORY_DAYS" envDefault:"7"`
	UploadDir      string        `env:"UPLOAD_DIR" envDefault:"/data/uploads"`
	UploadMaxBytes int64         `env:"UPLOAD_MAX_BYTES" envDefault:"5242880"`
	CORSOrigins    []string      `env:"CORS_ORIGINS" envDefault:"http://localhost:5173" envSeparator:","`
	TrustedProxies []string      `env:"TRUSTED_PROXIES" envSeparator:","` // e.g. the web container, so ClientIP() is the real client
	LogLevel       string        `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if cfg.HistoryDays < 1 || cfg.HistoryDays > 365 {
		return Config{}, errors.New("HISTORY_DAYS must be between 1 and 365")
	}
	return cfg, nil
}

func (c Config) IsProduction() bool { return c.AppEnv == "production" }
