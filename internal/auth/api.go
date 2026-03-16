package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	LicenseAPIURL    = "https://api.refactorkit.dev/v1/license/validate"
	CacheFileName    = ".refactorkit-license-cache.json"
	OfflineGraceDays = 7
)

type cachedLicense struct {
	Status   LicenseStatus `json:"status"`
	CachedAt time.Time     `json:"cached_at"`
}

// ValidateLicense checks a license key against the API, with offline caching
func ValidateLicense(key string) (*LicenseStatus, error) {
	if !ValidateKeyFormat(key) {
		return nil, fmt.Errorf("invalid license key format")
	}

	// Try API first
	status, err := callAPI(key)
	if err == nil {
		// Cache the result for offline use
		cacheResult(status)
		return status, nil
	}

	// API unreachable -- check cache
	cached, cacheErr := loadCache()
	if cacheErr == nil && time.Since(cached.CachedAt) < OfflineGraceDays*24*time.Hour {
		return &cached.Status, nil
	}

	return nil, fmt.Errorf("unable to validate license: %w", err)
}

func callAPI(key string) (*LicenseStatus, error) {
	body, _ := json.Marshal(map[string]string{"key": key})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(LicenseAPIURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var status LicenseStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

func cacheResult(status *LicenseStatus) {
	cacheDir, _ := os.UserCacheDir()
	path := filepath.Join(cacheDir, "refactorkit", CacheFileName)
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(cachedLicense{Status: *status, CachedAt: time.Now()})
	os.WriteFile(path, data, 0600)
}

func loadCache() (*cachedLicense, error) {
	cacheDir, _ := os.UserCacheDir()
	path := filepath.Join(cacheDir, "refactorkit", CacheFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cached cachedLicense
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}
