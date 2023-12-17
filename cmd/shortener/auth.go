package main

import (
	"errors"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/session"
	"github.com/KonBal/url-shortener/internal/app/user"
)

type authenticator interface {
	Authenticate(token string) (*user.User, error)
	Sign(token string) (string, error)
}

type authHandler struct {
	next          http.Handler
	authenticator authenticator
}

// AuthHandler create AuthHandler.
func AuthHandler(a authenticator) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &authHandler{
			next:          h,
			authenticator: a,
		}
	}
}

const authCookieKey = "user_id"

// ServeHTTP adds authentication to the pipeline.
func (h *authHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie(authCookieKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	u, err := h.authenticator.Authenticate(c.Value)
	if err != nil {
		if errors.Is(err, user.ErrAuthenticationFailed) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	s := &session.Session{UserID: u.UserID}

	ctx := session.ContextWithSession(req.Context(), s)
	req = req.WithContext(ctx)

	h.next.ServeHTTP(w, req)
}

type userStore interface {
	NewAnonymousUser() *user.User
}
type authenticationHandler struct {
	next          http.Handler
	authenticator authenticator
	userStore     userStore
}

// AuthenticationHandler creates handler that adds authentication to pipeline.
func AuthenticationHandler(a authenticator, s userStore) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &authenticationHandler{
			next:          h,
			authenticator: a,
			userStore:     s,
		}
	}
}

// ServeHTTP adds authentication to the pipeline. Creates new user if user is unauthenticated.
func (h *authenticationHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var u *user.User
	auth := true

	c, err := req.Cookie(authCookieKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			u = h.userStore.NewAnonymousUser()
			auth = false
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if auth {
		if u, err = h.authenticator.Authenticate(c.Value); err != nil {
			if errors.Is(err, user.ErrAuthenticationFailed) {
				u = h.userStore.NewAnonymousUser()
				auth = false
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	}

	if !auth {
		signed, err := h.authenticator.Sign(u.UserID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{Name: authCookieKey, Value: signed})
	}

	s := &session.Session{UserID: u.UserID}

	ctx := session.ContextWithSession(req.Context(), s)
	req = req.WithContext(ctx)

	h.next.ServeHTTP(w, req)
}
