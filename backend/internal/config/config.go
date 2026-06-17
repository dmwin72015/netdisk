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

	// Sensitive fields are sourced from environment variables so the on-disk
	// config can be committed without leaking secrets. Non-sensitive fields
	// (URLs, ports, scopes, paths, ...) stay in YAML where they are easy to
	// review. Naming convention:
	//   db.dsn                                  -> NETDISK_DB_DSN
	//   redis.password                          -> NETDISK_REDIS_PASSWORD
	//   jwt.secret                              -> NETDISK_JWT_SECRET
	//   oauth2.redirect_base_url                 -> NETDISK_OAUTH2_REDIRECT_BASE_URL
	//   oauth2.frontend_url                      -> NETDISK_OAUTH2_FRONTEND_URL
	//   oauth2.providers.<name>.client_id       -> NETDISK_OAUTH2_<NAME>_CLIENT_ID
	//   oauth2.providers.<name>.client_secret   -> NETDISK_OAUTH2_<NAME>_CLIENT_SECRET
	// AutomaticEnv() does not look up the map-typed oauth2.providers tree, so
	// each provider key is bound explicitly below.
	mustBind(v, "db.dsn", "NETDISK_DB_DSN")
	mustBind(v, "redis.password", "NETDISK_REDIS_PASSWORD")
	mustBind(v, "jwt.secret", "NETDISK_JWT_SECRET")
	mustBind(v, "oauth2.redirect_base_url", "NETDISK_OAUTH2_REDIRECT_BASE_URL")
	mustBind(v, "oauth2.frontend_url", "NETDISK_OAUTH2_FRONTEND_URL")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	bindOAuthProviderEnv(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	normalizePaths(path, cfg)

	if err := validateSecrets(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func mustBind(v *viper.Viper, key, env string) {
	// viper's BindEnv only fails when no key/env is provided.
	_ = v.BindEnv(key, env)
}

// bindOAuthProviderEnv wires every provider declared in the YAML to its
// matching NETDISK_OAUTH2_<NAME>_CLIENT_{ID,SECRET} env var. It must run after
// ReadInConfig so the provider names are known.
func bindOAuthProviderEnv(v *viper.Viper) {
	providers := v.GetStringMap("oauth2.providers")
	for name := range providers {
		envName := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(name))
		mustBind(v,
			fmt.Sprintf("oauth2.providers.%s.client_id", name),
			fmt.Sprintf("NETDISK_OAUTH2_%s_CLIENT_ID", envName),
		)
		mustBind(v,
			fmt.Sprintf("oauth2.providers.%s.client_secret", name),
			fmt.Sprintf("NETDISK_OAUTH2_%s_CLIENT_SECRET", envName),
		)
	}
}

// validateSecrets fails fast when a secret is still empty after env+yaml are
// merged. This catches "forgot to set NETDISK_JWT_SECRET in production" early
// instead of letting the server boot with a blank-string signing key.
func validateSecrets(cfg *Config) error {
	var missing []string
	if strings.TrimSpace(cfg.DB.DSN) == "" {
		missing = append(missing, "db.dsn (NETDISK_DB_DSN)")
	}
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		missing = append(missing, "jwt.secret (NETDISK_JWT_SECRET)")
	}
	for name, p := range cfg.OAuth2.Providers {
		// A provider stays optional until either id or secret is set; if one
		// is set, the other must be set too.
		hasID := strings.TrimSpace(p.ClientID) != ""
		hasSecret := strings.TrimSpace(p.ClientSecret) != ""
		if hasID != hasSecret {
			envBase := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(name))
			if !hasID {
				missing = append(missing, fmt.Sprintf("oauth2.providers.%s.client_id (NETDISK_OAUTH2_%s_CLIENT_ID)", name, envBase))
			}
			if !hasSecret {
				missing = append(missing, fmt.Sprintf("oauth2.providers.%s.client_secret (NETDISK_OAUTH2_%s_CLIENT_SECRET)", name, envBase))
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required secret(s) — set the environment variable(s): %s", strings.Join(missing, ", "))
	}
	return nil
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
