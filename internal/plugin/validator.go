package plugin

import (
	"errors"
	"strings"
)

var (
	ErrMissingName  = errors.New("plugin name is required")
	ErrMissingType  = errors.New("plugin type is required")
	ErrMissingEntry = errors.New("plugin entry is required")
)

// ValidateManifest performs static validation without executing plugin code.
func ValidateManifest(m Manifest) error {
	if strings.TrimSpace(m.Name) == "" {
		return ErrMissingName
	}
	if strings.TrimSpace(m.Type) == "" {
		return ErrMissingType
	}
	if strings.TrimSpace(m.Entry) == "" {
		return ErrMissingEntry
	}
	return nil
}
