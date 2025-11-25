package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

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
	PrivateKey       string `yaml:"private_key" env:"JWT_PRIVATE_KEY,required"`
	PublicKey        string `yaml:"public_key" env:"JWT_PUBLIC_KEY,required"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLHours  int    `yaml:"refresh_ttl_hours"`
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
	var cfg *Config

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			cfg = &Config{}
			yaml.Unmarshal(data, cfg)
			break
		}
	}

	if cfg == nil {
		panic("config file not found")
	}

	cfg.JWT.PrivateKey = os.Getenv("JWT_PRIVATE_KEY")
	cfg.JWT.PublicKey = os.Getenv("JWT_PUBLIC_KEY")

	cfg.JWT.PrivateKey = strings.ReplaceAll(cfg.JWT.PrivateKey, `\n`, "\n")
	cfg.JWT.PublicKey = strings.ReplaceAll(cfg.JWT.PublicKey, `\n`, "\n")

	if cfg.JWT.PrivateKey == "" {
		panic("JWT_PRIVATE_KEY not found in docker environment")
	}
	if cfg.JWT.PublicKey == "" {
		panic("JWT_PUBLIC_KEY not found in docker environment")
	}

	return cfg
}

func (c *Config) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	const op = "config.GetPrivateKey"

	block, _ := pem.Decode([]byte(c.JWT.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("%s: failed to parse PEM block", op)
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: parse EC private key: %w", op, err)
	}

	return privateKey, nil
}

func (c *Config) GetPublicKey() (*ecdsa.PublicKey, error) {
	const op = "config.GetPublicKey"

	block, _ := pem.Decode([]byte(c.JWT.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("%s: failed to parse PEM block", op)
	}

	genericPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: parse public key: %w", op, err)
	}

	publicKey, ok := genericPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%s: not an ECDSA public key", op)
	}

	return publicKey, nil
}

// func (c *Config) GetPrivateKey() (*ecdsa.PrivateKey, error) {
// 	const op = "config.GetPrivateKey"

// 	block, _ := pem.Decode([]byte(c.JWT.PrivateKey))
// 	if block == nil {
// 		return nil, fmt.Errorf("%s: failed to parse PEM block", op)
// 	}

// 	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: parse EC private key: %w", op, err)
// 	}

// 	return privateKey, nil
// }

// func (c *Config) GetPublicKey() (*ecdsa.PublicKey, error) {
// 	const op = "config.GetPublicKey"

// 	block, _ := pem.Decode([]byte(c.JWT.PublicKey))
// 	if block == nil {
// 		return nil, fmt.Errorf("%s: failed to parse PEM block", op)
// 	}

// 	genericPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: parse public key: %w", op, err)
// 	}

// 	publicKey, ok := genericPublicKey.(*ecdsa.PublicKey)
// 	if !ok {
// 		return nil, fmt.Errorf("%s: not an ECDSA public key", op)
// 	}
//
// 	return publicKey, nil
// }
