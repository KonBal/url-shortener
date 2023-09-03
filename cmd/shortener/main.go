package main

import (
	"fmt"
	"net/http"

	"github.com/KonBal/url-shortener/internal/app/base62"
	idgenerator "github.com/KonBal/url-shortener/internal/app/idgen"
	"github.com/KonBal/url-shortener/internal/app/operation"
	"github.com/KonBal/url-shortener/internal/app/storage"
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

	mux := http.NewServeMux()

	c := base62.Encoder{}
	s := storage.NewInMemory()
	idgen := idgenerator.New()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var o http.Handler

		switch r.Method {
		case http.MethodGet:
			o = operation.Expand{
				Service: operation.Expander{
					Decoder: c,
					Storage: s,
				},
			}
		case http.MethodPost:
			o = operation.Shorten{
				Service: operation.Shortener{
					Encoder:      c,
					Storage:      s,
					ShortURLHost: h,
					IDGen:        idgen,
				},
			}
		}

		if o != nil {
			o.ServeHTTP(w, r)
		}
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
