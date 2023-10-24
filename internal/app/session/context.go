package session

import "context"

type sessionContextKey struct{}

func FromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey{}).(*Session)
	return s
}

func ContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}
