package cfg

import "time"

type Logger struct {
	Level  string `env:"LOG_LEVEL, default=debug"`
	Format string `env:"LOG_FORMAT, default=text"`
}

type DataBase struct {
	User         string `env:"DB_USER"`
	Password     string `env:"DB_PASSWORD"`
	Port         string `env:"DB_PORT, default=5432"`
	Host         string `env:"DB_HOST"`
	DatabaseName string `env:"DB_NAME"`
}

type Keycloak struct {
	AuthURL          string `env:"AUTH_URL"`
	AuthRealm        string `env:"AUTH_REALM"`
	AuthClientID     string `env:"AUTH_CLIENT_ID"`
	AuthClientSecret string `env:"AUTH_CLIENT_SECRET"`
}

type Server struct {
	APIPath             string        `env:"SERVER_API_PATH, default=api"`
	Port                string        `env:"SERVER_PORT, default=8080"`
	WriteTimeout        time.Duration `env:"SERVER_WRITE_TIMEOUT, default=15s"`
	ReadTimeout         time.Duration `env:"SERVER_READ_TIMEOUT, default=15s"`
	IdleTimeout         time.Duration `env:"SERVER_IDLE_TIMEOUT, default=60s"`
	DeadlineOnInterrupt time.Duration `env:"SERVER_DEADLINE_ON_INTERRUPT, default=15s"`
	MinPageSize         int           `env:"SERVER_MIN_PAGE_SIZE, default=10"`
	MaxPageSize         int           `env:"SERVER_MAX_PAGE_SIZE, default=500"`
}
