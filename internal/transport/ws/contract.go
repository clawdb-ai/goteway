package ws

import "github.com/mac/goteway/pkg/types"

// Connection defines the minimum websocket connection abstraction.
type Connection interface {
	ReadFrame() (types.Frame, error)
	WriteFrame(types.Frame) error
	Close() error
}

// Hub describes connection registration and event broadcast behavior.
type Hub interface {
	Register(Connection) (clientID string, err error)
	Unregister(clientID string) error
	Broadcast(event types.Frame) error
}
