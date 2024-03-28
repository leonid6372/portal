package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	//Env      string `yaml:"env" env-default:"local"`
	LogLVL     string `yaml:"log_lvl" env-default:"info"`
	SQLStorage `yaml:"sql_storage"`
	HTTPServer `yaml:"http_server"`
	TokenTTL   time.Duration `yaml:"token_ttl" end-default:2h"`
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

func MustLoad() *Config {
	//configPath := "C:/Users/Leonid/Desktop/portal/config/local.yaml"
	configPath, err := filepath.Abs("../portal/config/local.yaml") // Относительный путь
	if err != nil {
		log.Fatal("error while reading config_path")
	}
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

	return &cfg
}
