package session

import "context"

type sessionContextKey struct{}

// FromContext returns a session object stored in given context.
func FromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey{}).(*Session)
	return s
}

// ContextWithSession adds session s to the given context.
func ContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}
