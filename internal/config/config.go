package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	//Env      string `yaml:"env" env-default:"local"`
	LogLVL       string `yaml:"log_lvl" env-default:"info"`
	SQLStorage   `yaml:"sql_storage"`
	HTTPServer   `yaml:"http_server"`
	BearerServer `yaml:"bearer_server"`
}

type SQLStorage struct {
	SQLDriver string `yaml:"sql_driver" env-required:"true"`
	SQLInfo   string `yaml:"sql_info" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type BearerServer struct {
	SecretPath string `yaml:"secret_path" env-required:"true"`
	Secret     string
	TokenTTL   time.Duration `yaml:"token_ttl" end-default:"2h"`
}

func MustLoad() *Config {
	configPath := "C:/Users/Leonid/Desktop/portal/config/local.yaml"
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	// check if oauth secret file exists
	if _, err := os.Stat(cfg.BearerServer.SecretPath); os.IsNotExist(err) {
		log.Fatalf("oauth secret file does not exist: %s", cfg.BearerServer.SecretPath)
	}

	secret, err := os.ReadFile(cfg.BearerServer.SecretPath)
	if err != nil {
		log.Fatalf("failed to read secret key: %s", err)
	}

	cfg.BearerServer.Secret = string(secret)

	return &cfg
}
