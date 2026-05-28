package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host               string   `mapstructure:"host"  validate:"required,hostname|ip"`
	Port               int      `mapstructure:"port"  validate:"required,gt=0,lte=65535"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

type DatabaseConfig struct {
	URL          string        `mapstructure:"url"            validate:"required,url"`
	PoolSize     int           `mapstructure:"pool_size"      validate:"gte=1"`
	MinIdleConns int           `mapstructure:"min_idle_conns" validate:"gte=0"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"   validate:"required"`
}

type RedisConfig struct {
	URL          string        `mapstructure:"url"            validate:"required,url"`
	PoolSize     int           `mapstructure:"pool_size"      validate:"gte=1"`
	MinIdleConns int           `mapstructure:"min_idle_conns" validate:"gte=0"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"   validate:"required"`
}

type Timeouts struct {
	HTTPRequest time.Duration `mapstructure:"http_request" validate:"required"`
	DBQuery     time.Duration `mapstructure:"db_query"     validate:"required"`
	CacheOp     time.Duration `mapstructure:"cache_op"     validate:"required"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"      validate:"required,min=32"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"  validate:"required,gt=0"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl" validate:"required,gt=0"`
}

type LoggerConfig struct {
	Level   string `mapstructure:"level"   validate:"required"`
	Service string `mapstructure:"service" validate:"required"`
	Env     string `mapstructure:"env"     validate:"required"`
	Version string `mapstructure:"version" validate:"required"`
}

type SimulatorConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Enabled bool          `mapstructure:"enabled"`
}

type SchedulerConfig struct {
	Interval time.Duration `mapstructure:"interval"`
}

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Timeouts  Timeouts        `mapstructure:"timeouts"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Simulator SimulatorConfig `mapstructure:"simulator"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

func Load() (*Config, error) {
	v := viper.New()

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "./configs/config.yml"
	}
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return &cfg, nil
}
