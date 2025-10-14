package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Name     string `yaml:"name"`
	SSLModel string `yaml:"ssl_model"`
}

func (c *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s", c.Host, c.User, c.Name, c.Pass, c.SSLModel)
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type Config struct {
	EnvLevel     string `yaml:"env"`
	DBConfig     `yaml:"DB_INFO"`
	ServerConfig `yaml:"HTTP_SERVER_INFO"`
	FrontendURL  string `yaml:"FRONTEND_INFO"`
}

func init() {
	// Загружаем .env файл (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env файл не найден, используем системные переменные")
	} else {
		log.Println("✅ .env файл загружен")
	}
}

func MustConfig(path string) *Config {
	envPath := os.Getenv("CONFIG_PATH")
	log.Printf("env init")
	if envPath == "" {
		envPath = path
		if envPath == "" {
			log.Fatalf("config path is not set")
		}
	}
	log.Printf("env prev stat")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", envPath)
	}

	log.Printf("env prev read")
	var myConfig Config
	if err := cleanenv.ReadConfig(envPath, &myConfig); err != nil {
		log.Fatalf("failed read config - %s: %v", envPath, err)
	}
	log.Printf("env return config")

	return &myConfig
}
