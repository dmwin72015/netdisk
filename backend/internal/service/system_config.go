package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type ConfigValueType string

const (
	ConfigTypeBytes   ConfigValueType = "bytes"
	ConfigTypeNumber  ConfigValueType = "number"
	ConfigTypeString  ConfigValueType = "string"
	ConfigTypeBool    ConfigValueType = "bool"
)

type ConfigDef struct {
	Key         string          `json:"key"`
	Default     any             `json:"defaultValue"`
	Description string          `json:"description"`
	Type        ConfigValueType `json:"type"`
}

type ConfigItem struct {
	Key          string          `json:"key"`
	Value        any             `json:"value"`
	DefaultValue any             `json:"defaultValue"`
	Description  string          `json:"description"`
	Type         ConfigValueType `json:"type"`
}

type SystemConfigService struct {
	pg     *pgxpool.Pool
	logger zerolog.Logger
	cache  sync.Map
	defs   []ConfigDef
	defMap map[string]ConfigDef
}

func NewSystemConfigService(pg *pgxpool.Pool, logger zerolog.Logger) *SystemConfigService {
	defs := []ConfigDef{
		{Key: "max_upload_size", Default: int64(4294967296), Description: "Maximum upload file size", Type: ConfigTypeBytes},
		{Key: "chunk_size", Default: int64(4194304), Description: "Upload chunk size", Type: ConfigTypeBytes},
		{Key: "default_quota", Default: int64(536870912000), Description: "Default user storage quota", Type: ConfigTypeBytes},
		{Key: "avatar_max_size", Default: int64(2097152), Description: "Maximum avatar file size", Type: ConfigTypeBytes},
		{Key: "retention_days", Default: float64(30), Description: "Trash retention days", Type: ConfigTypeNumber},
		{Key: "access_ttl_min", Default: float64(60), Description: "JWT access token TTL (minutes)", Type: ConfigTypeNumber},
		{Key: "refresh_ttl_hour", Default: float64(168), Description: "JWT refresh token TTL (hours)", Type: ConfigTypeNumber},
		{Key: "task_expiry_days", Default: float64(30), Description: "Upload task expiry days", Type: ConfigTypeNumber},
		{Key: "api_requests_per_min", Default: float64(600), Description: "API rate limit per minute", Type: ConfigTypeNumber},
	}

	defMap := make(map[string]ConfigDef, len(defs))
	for _, d := range defs {
		defMap[d.Key] = d
	}

	return &SystemConfigService{
		pg:     pg,
		logger: logger,
		defs:   defs,
		defMap: defMap,
	}
}

func (s *SystemConfigService) Definitions() []ConfigDef {
	return s.defs
}

func (s *SystemConfigService) Load(ctx context.Context) error {
	rows, err := s.pg.Query(ctx, "SELECT key, value FROM system_configs")
	if err != nil {
		return fmt.Errorf("load system configs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return fmt.Errorf("scan system config: %w", err)
		}
		var val any
		if err := json.Unmarshal(raw, &val); err != nil {
			s.logger.Warn().Err(err).Str("key", key).Msg("skip invalid config value")
			continue
		}
		s.cache.Store(key, val)
	}
	return rows.Err()
}

func (s *SystemConfigService) List(ctx context.Context) ([]ConfigItem, error) {
	items := make([]ConfigItem, 0, len(s.defs))
	for _, d := range s.defs {
		item := ConfigItem{
			Key:          d.Key,
			DefaultValue: d.Default,
			Description:  d.Description,
			Type:         d.Type,
		}
		if v, ok := s.cache.Load(d.Key); ok {
			item.Value = v
		} else {
			item.Value = d.Default
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *SystemConfigService) Get(key string) (any, bool) {
	if v, ok := s.cache.Load(key); ok {
		return v, true
	}
	def, ok := s.defMap[key]
	if ok {
		return def.Default, true
	}
	return nil, false
}

func (s *SystemConfigService) Set(ctx context.Context, key string, value any) error {
	if _, ok := s.defMap[key]; !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal config value: %w", err)
	}

	_, err = s.pg.Exec(ctx, `
		INSERT INTO system_configs (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, key, data)
	if err != nil {
		return fmt.Errorf("upsert system config: %w", err)
	}

	s.cache.Store(key, value)
	return nil
}

func (s *SystemConfigService) SetBatch(ctx context.Context, kv map[string]any) error {
	for k, v := range kv {
		if err := s.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigService) Reset(ctx context.Context, key string) error {
	if _, ok := s.defMap[key]; !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}

	_, err := s.pg.Exec(ctx, "DELETE FROM system_configs WHERE key = $1", key)
	if err != nil {
		return fmt.Errorf("delete system config: %w", err)
	}

	s.cache.Delete(key)
	return nil
}

func (s *SystemConfigService) ResetAll(ctx context.Context) error {
	_, err := s.pg.Exec(ctx, "DELETE FROM system_configs")
	if err != nil {
		return fmt.Errorf("delete all system configs: %w", err)
	}

	s.cache.Range(func(key, _ any) bool {
		s.cache.Delete(key)
		return true
	})
	return nil
}
