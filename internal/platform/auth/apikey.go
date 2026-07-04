package auth

import (
	"crypto/subtle"
	"strings"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

// APIKeyValidator checks X-API-Key headers against configured keys.
type APIKeyValidator struct {
	enabled bool
	keys    map[string]struct{}
}

// NewAPIKeyValidator constructs a validator. When disabled, all requests pass.
func NewAPIKeyValidator(enabled bool, keys []string) *APIKeyValidator {
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			set[k] = struct{}{}
		}
	}
	return &APIKeyValidator{enabled: enabled, keys: set}
}

// Validate returns an error if the key is missing or invalid.
func (v *APIKeyValidator) Validate(rawKey string) error {
	if !v.enabled {
		return nil
	}
	if rawKey == "" {
		return shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "missing API key")
	}
	for key := range v.keys {
		if subtle.ConstantTimeCompare([]byte(rawKey), []byte(key)) == 1 {
			return nil
		}
	}
	return shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "invalid API key")
}

// Enabled reports whether auth is enforced.
func (v *APIKeyValidator) Enabled() bool { return v.enabled }
