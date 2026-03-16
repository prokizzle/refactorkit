package auth

import (
	"strings"
	"time"
)

type Tier string

const (
	TierFree        Tier = "free"
	TierStandard    Tier = "standard"
	TierPremium     Tier = "premium"
	TierPremiumPlus Tier = "premium_plus"
)

type LicenseStatus struct {
	Valid             bool      `json:"valid"`
	Tier              Tier      `json:"tier"`
	ExpiresAt         time.Time `json:"expires_at"`
	ProjectsRemaining int       `json:"projects_remaining"`
	Email             string    `json:"email"`
}

// ValidateKeyFormat checks if a license key has the correct prefix format
// Format: rfkt_live_<32chars> or rfkt_test_<32chars>
func ValidateKeyFormat(key string) bool {
	if len(key) < 15 {
		return false
	}
	return strings.HasPrefix(key, "rfkt_live_") || strings.HasPrefix(key, "rfkt_test_")
}

// IsFreeTier returns true if no license or free tier
func (ls *LicenseStatus) IsFreeTier() bool {
	return ls == nil || ls.Tier == TierFree
}

// CanRefactor returns true if the license allows full refactoring
func (ls *LicenseStatus) CanRefactor() bool {
	return ls.Valid && ls.Tier != TierFree && ls.ProjectsRemaining > 0
}
