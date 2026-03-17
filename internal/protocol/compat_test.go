package protocol

import (
	"errors"
	"testing"

	"github.com/mac/goteway/pkg/types"
)

func TestNegotiateVersion(t *testing.T) {
	v, err := NegotiateVersion(1, 3, 1, 4)
	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	if v != 3 {
		t.Fatalf("expected version 3, got %d", v)
	}
}

func TestNegotiateVersionNoOverlap(t *testing.T) {
	_, err := NegotiateVersion(4, 5, 1, 3)
	if err == nil {
		t.Fatal("expected negotiation error")
	}
}

func TestValidateConnectParams(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := ValidateConnectParams(types.ConnectParams{
			MinProtocol: 1,
			MaxProtocol: 3,
			Client:      map[string]any{"name": "test"},
		})
		if err != nil {
			t.Fatalf("expected valid params, got err=%v", err)
		}
	})

	t.Run("invalid protocol range", func(t *testing.T) {
		err := ValidateConnectParams(types.ConnectParams{
			MinProtocol: 0,
			MaxProtocol: 3,
			Client:      map[string]any{"name": "test"},
		})
		if !errors.Is(err, ErrInvalidConnectProtocolRange) {
			t.Fatalf("expected ErrInvalidConnectProtocolRange, got %v", err)
		}
	})

	t.Run("missing client", func(t *testing.T) {
		err := ValidateConnectParams(types.ConnectParams{
			MinProtocol: 1,
			MaxProtocol: 3,
		})
		if !errors.Is(err, ErrMissingClientInfo) {
			t.Fatalf("expected ErrMissingClientInfo, got %v", err)
		}
	})
}
