package operation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Ping struct {
	Log     *logger.Logger
	Storage storage.Storage
}

func (o *Ping) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()

	err := o.Storage.Ping(ctx)
	if err != nil {
		o.Log.RequestError(req, fmt.Errorf("failed to ping db: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
