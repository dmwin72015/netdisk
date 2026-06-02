package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	DB        DBConfig        `mapstructure:"db"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Storage   StorageConfig   `mapstructure:"storage"`
	FFmpeg    FFmpegConfig    `mapstructure:"ffmpeg"`
	Log       LogConfig       `mapstructure:"log"`
	Limits    LimitsConfig    `mapstructure:"limits"`
	Trash     TrashConfig     `mapstructure:"trash"`
	Upload    UploadConfig    `mapstructure:"upload"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Media     MediaConfig     `mapstructure:"media"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	OAuth2    OAuth2Config    `mapstructure:"oauth2"`
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
	Level     string `mapstructure:"level"`
	Output    string `mapstructure:"output"`
	FilePath  string `mapstructure:"file_path"`
	MaxSizeMB int    `mapstructure:"max_size_mb"`
}

type LimitsConfig struct {
	DefaultStorageQuota int64 `mapstructure:"default_storage_quota"`
	BcryptCost          int   `mapstructure:"bcrypt_cost"`
	AvatarMaxSize       int64 `mapstructure:"avatar_max_size"`
	ContentDetectBytes  int   `mapstructure:"content_detect_bytes"`
}

type TrashConfig struct {
	RetentionDays int           `mapstructure:"retention_days"`
	PollInterval  time.Duration `mapstructure:"poll_interval"`
}

type UploadConfig struct {
	ChunkSize      int32         `mapstructure:"chunk_size"`
	TaskExpiryDays int           `mapstructure:"task_expiry_days"`
	MergeLockTTL   time.Duration `mapstructure:"merge_lock_ttl"`
}

type CacheConfig struct {
	ChallengeTTL time.Duration `mapstructure:"challenge_ttl"`
	ChunksTTL    time.Duration `mapstructure:"chunks_ttl"`
	PreCacheTTL  time.Duration `mapstructure:"precache_ttl"`
}

type MediaConfig struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
	BatchSize    int           `mapstructure:"batch_size"`
}

type RateLimitConfig struct {
	APIRequestsPerMin  int `mapstructure:"api_requests_per_min"`
	AuthRequestsPerMin int `mapstructure:"auth_requests_per_min"`
}

type OAuth2Config struct {
	RedirectBaseURL string                     `mapstructure:"redirect_base_url"`
	FrontendURL     string                     `mapstructure:"frontend_url"`
	Providers       map[string]OAuth2ProviderConfig `mapstructure:"providers"`
}

type OAuth2ProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	UserInfoURL  string `mapstructure:"user_info_url"`
	Scope        string `mapstructure:"scope"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("NETDISK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output", "console")
	v.SetDefault("log.file_path", "./logs/server.log")
	v.SetDefault("log.max_size_mb", 5)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	normalizePaths(path, cfg)

	return cfg, nil
}

func normalizePaths(configPath string, cfg *Config) {
	if cfg.Storage.Root == "" || filepath.IsAbs(cfg.Storage.Root) {
		return
	}
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return
	}
	cfg.Storage.Root = filepath.Clean(filepath.Join(filepath.Dir(absConfigPath), cfg.Storage.Root))
}
