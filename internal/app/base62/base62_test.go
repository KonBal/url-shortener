package base62

import (
	"math"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := map[string]struct {
		val  uint64
		want string
	}{
		"empty": {
			val:  0,
			want: "",
		},
		"max_uint8": {
			val:  math.MaxUint8,
			want: "he",
		},
		"max_uint16": {
			val:  math.MaxUint16,
			want: "bdr",
		},
		"max_uint32": {
			val:  math.MaxUint32,
			want: "dmpPQe",
		},
		"max_uint64": {
			val:  math.MaxUint64,
			want: "pIrkgbKrQ8v",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Encoder{}
			if got := c.Encode(tt.val); got != tt.want {
				t.Errorf("Encoder.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	c := Encoder{}

	b.Run("max_uint8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c.Encode(math.MaxUint8)
		}
	})

	b.Run("max_uint64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c.Encode(math.MaxUint64)
		}
	})
}
