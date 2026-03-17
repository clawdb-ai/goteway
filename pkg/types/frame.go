package types

// FrameType represents supported protocol frame types.
type FrameType string

const (
	FrameReq   FrameType = "req"
	FrameRes   FrameType = "res"
	FrameEvent FrameType = "event"
)

// Frame is the unified wire payload for WebSocket messages.
type Frame struct {
	Type         FrameType      `json:"type"`
	ID           string         `json:"id,omitempty"`
	Method       string         `json:"method,omitempty"`
	Params       any            `json:"params,omitempty"`
	OK           *bool          `json:"ok,omitempty"`
	Payload      any            `json:"payload,omitempty"`
	Event        string         `json:"event,omitempty"`
	Seq          uint64         `json:"seq,omitempty"`
	StateVersion uint64         `json:"stateVersion,omitempty"`
	Meta         map[string]any `json:"meta,omitempty"`
}

// ConnectParams models the gateway connect handshake parameters.
type ConnectParams struct {
	MinProtocol int            `json:"minProtocol"`
	MaxProtocol int            `json:"maxProtocol"`
	Client      map[string]any `json:"client"`
	Auth        map[string]any `json:"auth"`
}

// ConnectResult is returned on successful connect.
type ConnectResult struct {
	Protocol int    `json:"protocol"`
	ClientID string `json:"clientId"`
}
