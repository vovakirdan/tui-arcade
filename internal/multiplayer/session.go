package multiplayer

import "sync"

// SessionHandle is the transport-neutral interface for communicating with a session.
// It allows the coordinator and matches to send events without depending on Wish/Bubble Tea.
type SessionHandle interface {
	// ID returns the unique session identifier.
	ID() SessionID

	// Send sends an event to the session asynchronously.
	// Must be non-blocking; implementations should use buffered channels.
	Send(evt SessionEvent)

	// Done returns a channel that closes when the session ends.
	Done() <-chan struct{}
}

// ChannelSession is a SessionHandle implementation using Go channels.
// Used by the TUI layer to bridge Bubble Tea sessions with the coordinator.
type ChannelSession struct {
	id       SessionID
	events   chan SessionEvent
	done     chan struct{}
	doneOnce sync.Once
}

// NewChannelSession creates a new channel-based session handle.
// eventBufferSize controls how many events can be buffered before dropping.
func NewChannelSession(id SessionID, eventBufferSize int) *ChannelSession {
	if eventBufferSize < 1 {
		eventBufferSize = 64 // Default buffer size
	}
	return &ChannelSession{
		id:     id,
		events: make(chan SessionEvent, eventBufferSize),
		done:   make(chan struct{}),
	}
}

// ID returns the session identifier.
func (s *ChannelSession) ID() SessionID {
	return s.id
}

// Send sends an event to the session.
// If the buffer is full, old events are dropped to prevent blocking.
func (s *ChannelSession) Send(evt SessionEvent) {
	select {
	case <-s.done:
		// Session is closed, don't send
		return
	default:
	}

	select {
	case s.events <- evt:
		// Event sent successfully
	default:
		// Buffer full, drop oldest and retry
		select {
		case <-s.events:
			// Dropped oldest
		default:
		}
		// Try again (best effort)
		select {
		case s.events <- evt:
		default:
		}
	}
}

// Events returns the channel to receive events from.
// The TUI layer reads from this channel.
func (s *ChannelSession) Events() <-chan SessionEvent {
	return s.events
}

// Done returns the done channel.
func (s *ChannelSession) Done() <-chan struct{} {
	return s.done
}

// Close marks the session as done.
// Safe to call multiple times.
func (s *ChannelSession) Close() {
	s.doneOnce.Do(func() {
		close(s.done)
	})
}

// SessionRegistry tracks active sessions.
// Thread-safe for concurrent access.
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[SessionID]SessionHandle
}

// NewSessionRegistry creates a new session registry.
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{
		sessions: make(map[SessionID]SessionHandle),
	}
}

// Register adds a session to the registry.
func (r *SessionRegistry) Register(session SessionHandle) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID()] = session
}

// Unregister removes a session from the registry.
func (r *SessionRegistry) Unregister(id SessionID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
}

// Get retrieves a session by ID.
func (r *SessionRegistry) Get(id SessionID) (SessionHandle, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	return s, ok
}

// Count returns the number of registered sessions.
func (r *SessionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}
