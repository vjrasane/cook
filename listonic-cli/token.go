package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type TokenCache struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func (tc *TokenCache) Valid() bool {
	return tc.AccessToken != "" && time.Now().Unix() < tc.ExpiresAt
}

func tokenCachePath() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "listonic", "token.json")
}

func loadTokenCache() (*TokenCache, error) {
	data, err := os.ReadFile(tokenCachePath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading token cache: %w", err)
	}

	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parsing token cache: %w", err)
	}
	return &cache, nil
}

func saveTokenCache(cache *TokenCache) error {
	path := tokenCachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("marshaling token cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token cache: %w", err)
	}
	return nil
}
