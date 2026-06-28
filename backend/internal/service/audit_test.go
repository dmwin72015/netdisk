package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestToText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  pgtype.Text
	}{
		{name: "non-empty string", input: "hello", want: pgtype.Text{String: "hello", Valid: true}},
		{name: "empty string", input: "", want: pgtype.Text{}},
		{name: "whitespace", input: "  ", want: pgtype.Text{String: "  ", Valid: true}},
		{name: "unicode", input: "中国", want: pgtype.Text{String: "中国", Valid: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toText(tt.input)
			if got != tt.want {
				t.Errorf("toText(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAuditActionConstants(t *testing.T) {
	// Verify action constants are distinct and non-empty
	actions := []string{
		ActionLogin,
		ActionRegister,
		ActionLogout,
		ActionOAuthLogin,
		ActionPasswordChange,
		ActionFileUpload,
		ActionFileDelete,
		ActionFileRename,
		ActionFileMove,
		ActionDirCreate,
		ActionDirLock,
		ActionDirUnlock,
		ActionShareCreate,
		ActionShareDelete,
		ActionAdminCreateUser,
		ActionAdminDeleteUser,
		ActionAdminUpdateRole,
	}

	seen := make(map[string]bool, len(actions))
	for _, a := range actions {
		if a == "" {
			t.Error("action constant is empty")
		}
		if seen[a] {
			t.Errorf("duplicate action constant: %q", a)
		}
		seen[a] = true
	}
}

func TestAuditLogInputFields(t *testing.T) {
	input := AuditLogInput{
		UserID:       42,
		Action:       ActionLogin,
		ResourceType: "file",
		ResourceName: "test.txt",
		IP:           "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		Extra:        map[string]any{"key": "value"},
	}
	if input.UserID != 42 {
		t.Errorf("UserID = %d, want 42", input.UserID)
	}
	if input.Action != ActionLogin {
		t.Errorf("Action = %q, want %q", input.Action, ActionLogin)
	}
	if input.Extra["key"] != "value" {
		t.Errorf("Extra[key] = %v, want 'value'", input.Extra["key"])
	}
}
