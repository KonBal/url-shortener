package base62

import (
	"strings"
)

// Encoder.
type Encoder struct{}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const length = uint64(len(alphabet))

// Encode encodes uint64 value into a string.
// It uses base62-alphabet which is base64-alphabet without symbols '+' and '/'.
func (c Encoder) Encode(val uint64) string {
	var b strings.Builder
	b.Grow(11)

	for ; val > 0; val = val / length {
		b.WriteByte(alphabet[(val % length)])
	}

	return b.String()
}
