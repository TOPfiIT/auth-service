package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"dev"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	JWT      JWT      `yaml:"jwt"`
	Postgres Postgres `yaml:"postgres"`
	Redis    Redis    `yaml:"redis"`
}

type JWT struct {
	AccessTTLMinutes int `yaml:"access_ttl_minutes"`
	RefreshTTLHours  int `yaml:"refresh_ttl_hours"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c *Postgres) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}

func (c *Redis) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func MustLoad() *Config {
	paths := []string{"local.yaml", "config/local.yaml", "/app/config/local.yaml"}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			cfg := &Config{}
			yaml.Unmarshal(data, cfg)
			return cfg
		}
	}
	panic("config not found")
}
