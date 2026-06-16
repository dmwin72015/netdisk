package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	DefaultDirectoryUnlockTTLHours = 2
	PermanentDirectoryUnlockTTL    = -1
)

type UserSettings struct {
	ShowSystemDirs          bool   `json:"showSystemDirs"`
	UploadConcurrency       int    `json:"uploadConcurrency"`
	DuplicateStrategy       string `json:"duplicateStrategy"`
	DirectoryUnlockTTLHours int    `json:"directoryUnlockTtlHours"`
}

func DefaultUserSettings() UserSettings {
	return UserSettings{
		ShowSystemDirs:          true,
		UploadConcurrency:       3,
		DuplicateStrategy:       "prompt",
		DirectoryUnlockTTLHours: DefaultDirectoryUnlockTTLHours,
	}
}

func NormalizeUserSettings(settings UserSettings) UserSettings {
	defaults := DefaultUserSettings()
	if settings.UploadConcurrency < 1 || settings.UploadConcurrency > 10 {
		settings.UploadConcurrency = defaults.UploadConcurrency
	}
	switch settings.DuplicateStrategy {
	case "prompt", "overwrite", "keep_both", "skip":
	default:
		settings.DuplicateStrategy = defaults.DuplicateStrategy
	}
	switch settings.DirectoryUnlockTTLHours {
	case 1, 2, 6, 24, PermanentDirectoryUnlockTTL:
	default:
		settings.DirectoryUnlockTTLHours = defaults.DirectoryUnlockTTLHours
	}
	return settings
}

func (s *UserService) GetSettings(ctx context.Context, userID int64) (UserSettings, error) {
	settings := DefaultUserSettings()
	var raw []byte
	err := s.pg.QueryRow(ctx, `SELECT settings FROM user_settings WHERE user_id = $1`, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return settings, nil
		}
		return settings, fmt.Errorf("get settings: %w", err)
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return DefaultUserSettings(), nil
		}
	}
	return NormalizeUserSettings(settings), nil
}

func (s *UserService) SaveSettings(ctx context.Context, userID int64, settings UserSettings) (UserSettings, error) {
	settings = NormalizeUserSettings(settings)
	raw, err := json.Marshal(settings)
	if err != nil {
		return settings, fmt.Errorf("marshal settings: %w", err)
	}
	_, err = s.pg.Exec(ctx, `
		INSERT INTO user_settings (user_id, settings, updated_at)
		VALUES ($1, $2::jsonb, NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET settings = EXCLUDED.settings, updated_at = NOW()
	`, userID, string(raw))
	if err != nil {
		return settings, fmt.Errorf("save settings: %w", err)
	}
	return settings, nil
}
