package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Name     string `yaml:"name"`
	SSLModel string `yaml:"ssl_model"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Config struct {
	EnvLevel     string `yaml:"env"`
	DBConfig     `yaml:"DB_INFO"`
	ServerConfig `yaml:"HTTP_SERVER_INFO"`
}

func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func MustConfig(path string) *Config {
	envPath := os.Getenv("CONFIG_PATH")
	if envPath == "" {
		envPath = path
		if envPath == "" {
			log.Fatalf("config path is not set")
		}
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", envPath)
	}

	var myConfig *Config
	if err := cleanenv.ReadConfig(envPath, myConfig); err != nil {
		log.Fatalf("failed read config - %s: %v", envPath, err)
	}

	return myConfig
}
