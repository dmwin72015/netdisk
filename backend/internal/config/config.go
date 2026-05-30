package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	DB      DBConfig      `mapstructure:"db"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Storage StorageConfig `mapstructure:"storage"`
	FFmpeg  FFmpegConfig  `mapstructure:"ffmpeg"`
	Log     LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	Port            int      `mapstructure:"port"`
	ReadTimeoutSec  int      `mapstructure:"read_timeout_sec"`
	WriteTimeoutSec int      `mapstructure:"write_timeout_sec"`
	CORSOrigins     []string `mapstructure:"cors_origins"`
}

type DBConfig struct {
	DSN      string `mapstructure:"dsn"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	AccessTTLMin   int    `mapstructure:"access_ttl_min"`
	RefreshTTLHour int    `mapstructure:"refresh_ttl_hour"`
}

type StorageConfig struct {
	Root          string `mapstructure:"root"`
	MaxUploadSize int64  `mapstructure:"max_upload_size"`
	TmpDir        string `mapstructure:"tmp_dir"`
	FilesDir      string `mapstructure:"files_dir"`
	AvatarsDir    string `mapstructure:"avatars_dir"`
	HLSDir        string `mapstructure:"hls_dir"`
}

type FFmpegConfig struct {
	Path        string `mapstructure:"path"`
	FFprobePath string `mapstructure:"ffprobe_path"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("NETDISK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
