package auth

import "testing"

func TestValidateTokenSuccess(t *testing.T) {
	svc := New("secret")
	err := svc.Validate(map[string]any{"type": "token", "token": "secret"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidateTokenFailure(t *testing.T) {
	svc := New("secret")
	err := svc.Validate(map[string]any{"type": "token", "token": "bad"})
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}
