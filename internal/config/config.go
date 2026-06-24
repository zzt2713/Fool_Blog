package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Addr string `yaml:"addr"`
	Name string `yaml:"name"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Path     string `yaml:"path"`
	DSN      string `yaml:"dsn"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type SecurityConfig struct {
	SessionKey   string `yaml:"session_key"`
	PasswordSalt string `yaml:"password_salt"`
}

type UploadConfig struct {
	MaxSizeMB  int    `yaml:"max_size_mb"`
	AvatarDir  string `yaml:"avatar_dir"`
	CoverDir   string `yaml:"cover_dir"`
	ArticleDir string `yaml:"article_dir"`
}

type WallpaperConfig struct {
	Enabled   bool   `yaml:"enabled"`
	API       string `yaml:"api"`
	CustomURL string `yaml:"custom_url"`
}

type AdminConfig struct {
	DefaultUsername string `yaml:"default_username"`
	DefaultPassword string `yaml:"default_password"`
}

type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type AIConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

type Config struct {
	App       AppConfig       `yaml:"app"`
	Database  DatabaseConfig  `yaml:"database"`
	Security  SecurityConfig  `yaml:"security"`
	Upload    UploadConfig    `yaml:"upload"`
	Wallpaper WallpaperConfig `yaml:"wallpaper"`
	Admin     AdminConfig     `yaml:"admin"`
	Email     EmailConfig     `yaml:"email"`
	AI        AIConfig        `yaml:"ai"`
}

var Current *Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	Current = c
	return c, nil
}
