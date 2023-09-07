package config

import "flag"

type Options struct {
	HostAddress         string
	ShortURLHostAddress string
}

var opt Options

func Parse() {
	flag.StringVar(&opt.HostAddress, "a", "localhost:8080", "host address")
	flag.StringVar(&opt.ShortURLHostAddress, "b", "localhost:8080", "address of short url host")

	flag.Parse()
}

func Get() Options {
	return opt
}
