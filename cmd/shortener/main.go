package main

import (
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/base62"
	idgenerator "github.com/KonBal/url-shortener/internal/app/idgen"
	"github.com/KonBal/url-shortener/internal/app/operation"
	"github.com/KonBal/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type Host struct {
	addr string
	port int
}

func (h Host) Address() string {
	return fmt.Sprintf("%s:%d", h.addr, h.port)
}

func main() {
	h := Host{
		addr: `http://localhost`,
		port: 8080,
	}

	router := chi.NewRouter()

	c := base62.Encoder{}
	s := storage.NewInMemory()
	idgen := idgenerator.New()

	router.Post("/",
		operation.ShortenHandle(operation.Shortener{
			Encoder:      c,
			Storage:      s,
			ShortURLHost: h,
			IDGen:        idgen,
		}))

	router.Get("/{short}",
		operation.ExpandHandle(operation.Expander{
			Decoder: c,
			Storage: s,
		}))

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
