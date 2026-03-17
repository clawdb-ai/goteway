package plugin

import "errors"

var (
	ErrMissingName  = errors.New("plugin name is required")
	ErrMissingType  = errors.New("plugin type is required")
	ErrMissingEntry = errors.New("plugin entry is required")
)

// ValidateManifest performs static validation without executing plugin code.
func ValidateManifest(m Manifest) error {
	if m.Name == "" {
		return ErrMissingName
	}
	if m.Type == "" {
		return ErrMissingType
	}
	if m.Entry == "" {
		return ErrMissingEntry
	}
	return nil
}
