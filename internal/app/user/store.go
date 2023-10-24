package user

import (
	"fmt"
	"time"
)

type rand interface {
	Next() uint64
}

type User struct {
	UserID string
}

type Store struct {
	rand rand
}

func NewStore(rand rand) Store {
	return Store{rand: rand}
}

func (s Store) NewAnonymousUser() *User {
	return &User{UserID: s.generateID()}
}

func (s Store) generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().Unix(), s.rand.Next())
}
