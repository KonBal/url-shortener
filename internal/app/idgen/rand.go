// Module generates pseudo-random integers (uint64).
package idgen

import "math/rand"

type uint64Gen struct{}

// New created new generator.
func New() uint64Gen {
	return uint64Gen{}
}

// Next returns another random value.
func (g uint64Gen) Next() uint64 {
	return rand.Uint64()
}
