package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	Ext      ExtConfig      `yaml:"ext"`
}

type AppConfig struct {
	Name                      string
	Version                   string
	MaxHeaderBytes            int    `yaml:"max_header_bytes"`
	MaxMultipartMemory        int64  `yaml:"max_multipart_memory"`
	WebsocketMessageSizeLimit int    `yaml:"websocket_message_size_limit"`
	PublicDir                 string `yaml:"public_dir"`
	LogDir                    string `yaml:"log_dir"`
	UploadDir                 string `yaml:"upload_dir"`
	LogTimeFormat             string `yaml:"log_time_format"`
	JwtPrivateKeyPath         string `yaml:"jwt_private_key_path"`
	JwtPublicKeyPath          string `yaml:"jwt_public_key_path"`
}

type AuthConfig struct {
	GoogleFBPath   string `yaml:"google_fb_path"`
	GoogleId       string `yaml:"google_id"`
	GoogleSecret   string `yaml:"google_secret"`
	FacebookId     string `yaml:"facebook_id"`
	FacebookSecret string `yaml:"facebook_secret"`
}

type ExtConfig struct {
	MailDomainPath                  string `yaml:"mail_domain_path"`
	GeoIpPath                       string `yaml:"geo_ip_path"`
	ProfanityFilterPath             string `yaml:"profanity_filter_path"`
}

type ServerConfig struct {
	Scheme       string
	Host         string
	Port         int64
	Mode         string
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

func (s ServerConfig) GetFullHostName() string {
	return fmt.Sprintf("%s://%s:%d", s.Scheme, s.Host, s.Port)
}

type DatabaseConfig struct {
	Dialect     string
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Host        string `yaml:"host"`
	Name        string `yaml:"name"`
	Loc         string `yaml:"loc"`
	MaxIdle     int    `yaml:"max_idle"`
	MaxOpen     int    `yaml:"max_open"`
	TablePrefix string `yaml:"table_prefix"`
}

type RedisConfig struct {
	Host        string
	Password    string
	MaxIdle     int           `yaml:"max_idle"`
	MaxActive   int           `yaml:"max_active"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

var (
	Global   GlobalConfig
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Ext      ExtConfig
)

func Load(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &Global)
	if err != nil {
		return err
	}

	App = Global.App
	Server = Global.Server
	Database = Global.Database
	Redis = Global.Redis
	Auth = Global.Auth
	Ext = Global.Ext

	return nil
}

func Setup() {
	if os.Getenv("CONFIG_PATH") == "" {
		panic("Failed load config")
	}

	file := os.Getenv("CONFIG_PATH")

	err := Load(file)
	if err != nil {
		panic("Failed load config")
	}
}
