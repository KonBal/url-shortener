package base62

import (
	"errors"
	"math"
	"strings"
)

type Encoder struct{}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const length = uint64(len(alphabet))

func (c Encoder) Encode(val uint64) string {
	var b strings.Builder
	b.Grow(11)

	for ; val > 0; val = val / length {
		b.WriteByte(alphabet[(val % length)])
	}

	return b.String()
}

func (c Encoder) Decode(encoded string) (uint64, error) {
	var val uint64

	for i, ch := range encoded {
		pos := strings.IndexRune(alphabet, ch)

		if pos == -1 {
			return uint64(pos), errors.New("invalid character: " + string(ch))
		}

		val += uint64(pos) * uint64(math.Pow(float64(length), float64(i)))
	}

	return val, nil
}
