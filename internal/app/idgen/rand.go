package idgenerator

import "math/rand"

type uint64Gen struct{}

func New() uint64Gen {
	return uint64Gen{}
}

func (g uint64Gen) Next() uint64 {
	return rand.Uint64()
}
