package operation

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/KonBal/url-shortener/internal/app/logger"
)

type Ping struct {
	Log *logger.Logger
	DB  *sql.DB
}

func (o *Ping) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()

	err := o.DB.PingContext(ctx)
	if err != nil {
		o.Log.RequestError(req, fmt.Errorf("failed to ping db: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
