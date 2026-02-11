package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenCachePath(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "")
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".cache", "listonic", "token.json")
		if got := tokenCachePath(); got != want {
			t.Errorf("tokenCachePath() = %q, want %q", got, want)
		}
	})

	t.Run("xdg", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
		want := "/tmp/xdg-cache/listonic/token.json"
		if got := tokenCachePath(); got != want {
			t.Errorf("tokenCachePath() = %q, want %q", got, want)
		}
	})
}

func TestSaveAndLoadTokenCache(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	cache := &TokenCache{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now().Unix() + 3600,
	}

	if err := saveTokenCache(cache); err != nil {
		t.Fatalf("saveTokenCache() error = %v", err)
	}

	loaded, err := loadTokenCache()
	if err != nil {
		t.Fatalf("loadTokenCache() error = %v", err)
	}
	if loaded.AccessToken != cache.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, cache.AccessToken)
	}
	if loaded.RefreshToken != cache.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, cache.RefreshToken)
	}
	if loaded.ExpiresAt != cache.ExpiresAt {
		t.Errorf("ExpiresAt = %d, want %d", loaded.ExpiresAt, cache.ExpiresAt)
	}
}

func TestLoadTokenCacheMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	loaded, err := loadTokenCache()
	if err != nil {
		t.Fatalf("loadTokenCache() error = %v", err)
	}
	if loaded != nil {
		t.Errorf("loadTokenCache() = %v, want nil", loaded)
	}
}

func TestTokenCacheValid(t *testing.T) {
	valid := &TokenCache{
		AccessToken: "tok",
		ExpiresAt:   time.Now().Unix() + 3600,
	}
	if !valid.Valid() {
		t.Error("expected valid token to be valid")
	}

	expired := &TokenCache{
		AccessToken: "tok",
		ExpiresAt:   time.Now().Unix() - 1,
	}
	if expired.Valid() {
		t.Error("expected expired token to be invalid")
	}

	empty := &TokenCache{
		ExpiresAt: time.Now().Unix() + 3600,
	}
	if empty.Valid() {
		t.Error("expected empty access token to be invalid")
	}
}
