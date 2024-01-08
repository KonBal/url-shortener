package main

import "os"

func foo() {
	os.Exit(1)
}

func main() {
	f := func() {
		os.Exit(1)
	}

	f()
}
