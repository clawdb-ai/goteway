package protocol

import (
	"errors"

	"github.com/mac/goteway/pkg/types"
)

var ErrProtocolNegotiation = errors.New("protocol negotiation failed")

var (
	ErrInvalidConnectProtocolRange = errors.New("invalid connect protocol range")
	ErrMissingClientInfo           = errors.New("missing client info")
)

var (
	okTrue  = true
	okFalse = false
)

// NegotiateVersion resolves protocol using client min/max and server min/max.
func NegotiateVersion(clientMin, clientMax, serverMin, serverMax int) (int, error) {
	if clientMin > clientMax {
		return 0, ErrProtocolNegotiation
	}
	candidate := min(clientMax, serverMax)
	if candidate < max(clientMin, serverMin) {
		return 0, ErrProtocolNegotiation
	}
	return candidate, nil
}

// ValidateConnectParams validates wire-level connect parameters before negotiation.
func ValidateConnectParams(p types.ConnectParams) error {
	if p.MinProtocol <= 0 || p.MaxProtocol <= 0 || p.MinProtocol > p.MaxProtocol {
		return ErrInvalidConnectProtocolRange
	}
	if len(p.Client) == 0 {
		return ErrMissingClientInfo
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NewConnectSuccess returns a standard connect response payload.
func NewConnectSuccess(id string, protocol int, clientID string) types.Frame {
	return types.Frame{
		Type:    types.FrameRes,
		ID:      id,
		OK:      &okTrue,
		Payload: types.ConnectResult{Protocol: protocol, ClientID: clientID},
	}
}

// NewErrorResponse returns a standard error response payload.
func NewErrorResponse(id string, code, message string) types.Frame {
	return types.Frame{
		Type: types.FrameRes,
		ID:   id,
		OK:   &okFalse,
		Payload: map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		},
	}
}
