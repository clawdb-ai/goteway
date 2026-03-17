package runtime

import (
	"github.com/mac/goteway/internal/apperr"
	"github.com/mac/goteway/internal/auth"
	"github.com/mac/goteway/internal/config"
	"github.com/mac/goteway/internal/protocol"
	"github.com/mac/goteway/internal/session"
	"github.com/mac/goteway/pkg/types"
)

// Service hosts protocol-compatible business logic.
type Service struct {
	cfg     config.Config
	auth    *auth.Service
	session *session.Manager
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:     cfg,
		auth:    auth.New(cfg.AuthToken),
		session: session.NewManager(),
	}
}

// Connect implements protocol handshake semantics.
func (s *Service) Connect(reqID string, p types.ConnectParams) types.Frame {
	if err := s.auth.Validate(p.Auth); err != nil {
		return protocol.NewErrorResponse(reqID, "ERR_UNAUTHORIZED", "authentication failed")
	}

	v, err := protocol.NegotiateVersion(p.MinProtocol, p.MaxProtocol, s.cfg.ProtocolMin, s.cfg.ProtocolMax)
	if err != nil {
		return protocol.NewErrorResponse(reqID, "ERR_PROTOCOL_NEGOTIATION", "no compatible protocol version")
	}

	clientID := s.session.NewClientID()
	s.session.Put(clientID, map[string]any{
		"client": p.Client,
	})
	return protocol.NewConnectSuccess(reqID, v, clientID)
}

func (s *Service) ChatCompletions(_ map[string]any) (map[string]any, error) {
	return nil, apperr.ErrNotImplemented
}

func (s *Service) InvokeTool(_ map[string]any) (map[string]any, error) {
	return nil, apperr.ErrNotImplemented
}
