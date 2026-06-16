package jwtutil

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	return NewManager("secret-for-test", 15*time.Minute, 7*24*time.Hour)
}

func TestAccessTokenRoundtrip(t *testing.T) {
	m := newTestManager(t)
	var uid int64 = 4711
	tok, exp, err := m.GenerateAccessToken(uid, "session-1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if exp.Before(time.Now()) {
		t.Fatalf("expected future expiry, got %v", exp)
	}
	c, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if c.UserID != uid {
		t.Fatalf("uid mismatch: got %v want %v", c.UserID, uid)
	}
	if c.Type != TokenTypeAccess {
		t.Fatalf("type mismatch: got %q", c.Type)
	}
	if c.SID != "session-1" {
		t.Fatalf("sid mismatch: got %q", c.SID)
	}
	if c.Subject != "4711" {
		t.Fatalf("subject should match uid as string, got %q", c.Subject)
	}
}

func TestRefreshTokenIsRefreshType(t *testing.T) {
	m := newTestManager(t)
	tok, jti, _, err := m.GenerateRefreshToken(int64(1), "session-2")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if jti == "" {
		t.Fatalf("expected non-empty jti")
	}
	c, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if c.Type != TokenTypeRefresh {
		t.Fatalf("type mismatch: got %q", c.Type)
	}
	if c.SID != "session-2" {
		t.Fatalf("sid mismatch: got %q", c.SID)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	m := NewManager("k", -time.Minute, -time.Minute)
	tok, _, err := m.GenerateAccessToken(1, "session")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = m.Parse(tok)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}

func TestTamperedSignatureIsRejected(t *testing.T) {
	m := newTestManager(t)
	tok, _, err := m.GenerateAccessToken(1, "session")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(tok) == 0 {
		t.Fatalf("empty token")
	}
	bad := tok[:len(tok)-1] + flipLastChar(tok[len(tok)-1])
	if _, err := m.Parse(bad); err == nil {
		t.Fatalf("expected error for tampered token")
	}
}

func TestForeignSignerIsRejected(t *testing.T) {
	m := newTestManager(t)
	other := NewManager("different-secret", 15*time.Minute, 7*24*time.Hour)
	tok, _, err := other.GenerateAccessToken(1, "session")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.Parse(tok); err == nil {
		t.Fatalf("expected error for foreign signature")
	}
}

func flipLastChar(b byte) string {
	if b == 'A' {
		return "B"
	}
	return "A"
}

var _ jwt.Claims = (*Claims)(nil)
