package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	LinkedIn LinkedInConfig `mapstructure:"linkedin" validate:"required"`
	Browser  BrowserConfig  `mapstructure:"browser" validate:"required"`
	Search   SearchConfig   `mapstructure:"search" validate:"required"`
	DB       DBConfig       `mapstructure:"db" validate:"required"`
	Logging  LoggingConfig  `mapstructure:"logging" validate:"required"`
}

type LinkedInConfig struct {
	Email    string `mapstructure:"email" validate:"required,email"`
	Password string `mapstructure:"password" validate:"required"`
}

type BrowserConfig struct {
	Headless     bool   `mapstructure:"headless"`
	UserAgent    string `mapstructure:"user_agent" validate:"required"`
	SlowMoMillis int    `mapstructure:"slow_mo_ms" validate:"gte=0"`
	SessionDir   string `mapstructure:"session_dir" validate:"required"`
}

type DBConfig struct {
	Driver string `mapstructure:"driver" validate:"required,oneof=sqlite"`
	DSN    string `mapstructure:"dsn" validate:"required"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format string `mapstructure:"format" validate:"required,oneof=json console"`
}

type SearchConfig struct {
	Keywords   []string `mapstructure:"keywords" validate:"required,min=1"`
	Locations  []string `mapstructure:"locations"`
	Companies  []string `mapstructure:"companies"`
	JobTitles  []string `mapstructure:"job_titles"`
	MaxPages   int      `mapstructure:"max_pages" validate:"gte=1,lte=10"`
	PageDelayS int      `mapstructure:"page_delay_seconds" validate:"gte=2,lte=10"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	_ = godotenv.Load()
	v.SetEnvPrefix("SCRAPEDIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("linkedin.email")
	_ = v.BindEnv("linkedin.password")

	_ = v.BindEnv("browser.headless")
	_ = v.BindEnv("browser.user_agent")
	_ = v.BindEnv("browser.slow_mo_ms")
	_ = v.BindEnv("browser.session_dir")

	_ = v.BindEnv("search.keywords")
	_ = v.BindEnv("search.locations")
	_ = v.BindEnv("search.companies")
	_ = v.BindEnv("search.job_titles")
	_ = v.BindEnv("search.max_pages")
	_ = v.BindEnv("search.page_delay_seconds")

	_ = v.BindEnv("db.driver")
	_ = v.BindEnv("db.dsn")

	_ = v.BindEnv("logging.level")
	_ = v.BindEnv("logging.format")

	// Defaults (important for DX)
	v.SetDefault("browser.headless", false)
	v.SetDefault("browser.slow_mo_ms", 0)
	v.SetDefault("browser.session_dir", "data/browser")
	v.SetDefault("db.driver", "sqlite")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	fmt.Println("EMAIL:", v.GetString("linkedin.email"))
	fmt.Println("DSN:", v.GetString("db.dsn"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
