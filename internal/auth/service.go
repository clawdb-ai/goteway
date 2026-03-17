package auth

import (
	"errors"
)

var ErrUnauthorized = errors.New("unauthorized")

// Service validates handshake auth payloads.
type Service struct {
	token string
}

func New(token string) *Service {
	return &Service{token: token}
}

// Validate supports token auth shape:
// {"type":"token","token":"..."}
func (s *Service) Validate(auth map[string]any) error {
	if s.token == "" {
		return nil
	}
	t, _ := auth["type"].(string)
	if t != "token" {
		return ErrUnauthorized
	}
	v, _ := auth["token"].(string)
	if v == "" || v != s.token {
		return ErrUnauthorized
	}
	return nil
}
