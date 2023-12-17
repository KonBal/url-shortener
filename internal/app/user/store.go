package user

import (
	"fmt"
	"time"
)

type rand interface {
	Next() uint64
}

// Represents user for authentication.
type User struct {
	UserID string
}

// Store of users.
type Store struct {
	rand rand
}

// NewStore returns new store.
func NewStore(rand rand) Store {
	return Store{rand: rand}
}

// NewAnonymousUser returns a new user object with randomly generated ID.
func (s Store) NewAnonymousUser() *User {
	return &User{UserID: s.generateID()}
}

func (s Store) generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().Unix(), s.rand.Next())
}
