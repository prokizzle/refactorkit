package auth

import (
	"testing"
)

func TestValidateKeyFormat(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "valid live key",
			key:  "rfkt_live_abcdefghijklmnopqrstuvwxyz123456",
			want: true,
		},
		{
			name: "valid test key",
			key:  "rfkt_test_abcdefghijklmnopqrstuvwxyz123456",
			want: true,
		},
		{
			name: "invalid prefix",
			key:  "invalid_abcdefghijklmnopqrstuvwxyz123456",
			want: false,
		},
		{
			name: "empty string",
			key:  "",
			want: false,
		},
		{
			name: "too short",
			key:  "rfkt_live_ab",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateKeyFormat(tt.key)
			if got != tt.want {
				t.Errorf("ValidateKeyFormat(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestLicenseStatus_IsFreeTier(t *testing.T) {
	tests := []struct {
		name   string
		status *LicenseStatus
		want   bool
	}{
		{
			name:   "nil status returns true",
			status: nil,
			want:   true,
		},
		{
			name:   "free tier returns true",
			status: &LicenseStatus{Tier: TierFree},
			want:   true,
		},
		{
			name:   "standard tier returns false",
			status: &LicenseStatus{Tier: TierStandard},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsFreeTier()
			if got != tt.want {
				t.Errorf("IsFreeTier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseStatus_CanRefactor(t *testing.T) {
	tests := []struct {
		name   string
		status *LicenseStatus
		want   bool
	}{
		{
			name: "valid standard with projects returns true",
			status: &LicenseStatus{
				Valid:             true,
				Tier:              TierStandard,
				ProjectsRemaining: 5,
			},
			want: true,
		},
		{
			name: "free tier returns false",
			status: &LicenseStatus{
				Valid:             true,
				Tier:              TierFree,
				ProjectsRemaining: 5,
			},
			want: false,
		},
		{
			name: "zero projects returns false",
			status: &LicenseStatus{
				Valid:             true,
				Tier:              TierStandard,
				ProjectsRemaining: 0,
			},
			want: false,
		},
		{
			name: "invalid license returns false",
			status: &LicenseStatus{
				Valid:             false,
				Tier:              TierPremium,
				ProjectsRemaining: 10,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.CanRefactor()
			if got != tt.want {
				t.Errorf("CanRefactor() = %v, want %v", got, tt.want)
			}
		})
	}
}
