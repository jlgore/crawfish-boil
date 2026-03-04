package simulation

import (
	"crypto/rand"
	"fmt"
	"sync"
)

// Session tracks state for a connected client.
type Session struct {
	ID            string
	Authenticated bool
	ClientInfo    map[string]any
	AuthMode      string
	AuthValue     string
}

var (
	sessions sync.Map
)

// NewSession creates a tracked session.
func NewSession() *Session {
	s := &Session{
		ID:         generateID(),
		ClientInfo: make(map[string]any),
	}
	sessions.Store(s.ID, s)
	return s
}

// RemoveSession cleans up a session.
func RemoveSession(id string) {
	sessions.Delete(id)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
