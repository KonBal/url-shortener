package base62

import (
	"fmt"
	"math"
)

func Example() {
	en := Encoder{}

	fmt.Println(en.Encode(123))
	fmt.Println(en.Encode(math.MaxUint64))
}
