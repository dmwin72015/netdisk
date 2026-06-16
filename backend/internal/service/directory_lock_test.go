package service

import (
	"testing"
)

func TestDirectoryUnlockExpiry(t *testing.T) {
	tests := []struct {
		name        string
		ttlHours    int
		wantPermanent bool
	}{
		{name: "permanent", ttlHours: PermanentDirectoryUnlockTTL, wantPermanent: true},
		{name: "1 hour", ttlHours: 1, wantPermanent: false},
		{name: "2 hours", ttlHours: 2, wantPermanent: false},
		{name: "6 hours", ttlHours: 6, wantPermanent: false},
		{name: "24 hours", ttlHours: 24, wantPermanent: false},
		{name: "defaults to 2", ttlHours: 99, wantPermanent: false},
		{name: "zero defaults", ttlHours: 0, wantPermanent: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, permanent := directoryUnlockExpiry(tt.ttlHours)
			if permanent != tt.wantPermanent {
				t.Errorf("directoryUnlockExpiry(%d) permanent = %v, want %v", tt.ttlHours, permanent, tt.wantPermanent)
			}
		})
	}
}

func TestNormalizeUserSettings_UnlockTTL(t *testing.T) {
	tests := []struct {
		name string
		input int
		want int
	}{
		{name: "valid 1", input: 1, want: 1},
		{name: "valid 2", input: 2, want: 2},
		{name: "valid 6", input: 6, want: 6},
		{name: "valid 24", input: 24, want: 24},
		{name: "valid permanent", input: PermanentDirectoryUnlockTTL, want: PermanentDirectoryUnlockTTL},
		{name: "invalid defaults to 2", input: 3, want: DefaultDirectoryUnlockTTLHours},
		{name: "zero defaults", input: 0, want: DefaultDirectoryUnlockTTLHours},
		{name: "negative invalid", input: -5, want: DefaultDirectoryUnlockTTLHours},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := NormalizeUserSettings(UserSettings{
				DirectoryUnlockTTLHours: tt.input,
				UploadConcurrency:       3,
				DuplicateStrategy:       "prompt",
				ShowSystemDirs:          true,
			})
			if settings.DirectoryUnlockTTLHours != tt.want {
				t.Errorf("DirectoryUnlockTTLHours = %d, want %d", settings.DirectoryUnlockTTLHours, tt.want)
			}
		})
	}
}
