package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KonBal/url-shortener/internal/app/logger"
	"github.com/KonBal/url-shortener/internal/app/session"
	"github.com/KonBal/url-shortener/internal/app/storage"
)

type Delete struct {
	Log     *logger.Logger
	Service interface {
		Delete(ctx context.Context, userID string, urls []string) error
	}
}

func (o *Delete) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var urls []string

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		o.Log.RequestError(req, fmt.Errorf("read request body: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := req.Context()
	s := session.FromContext(ctx)

	if err := o.Service.Delete(ctx, s.UserID, urls); err != nil {
		o.Log.RequestError(req, err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type DeletionWorker struct {
	entriesCh  chan storage.EntryToDelete
	workPeriod time.Duration
	storage    storage.Storage
	log        *logger.Logger
}

func NewDeletionWorker(s storage.Storage, log *logger.Logger,
	bufSize, workPeriodSec int64) *DeletionWorker {
	w := &DeletionWorker{
		storage:    s,
		entriesCh:  make(chan storage.EntryToDelete, bufSize),
		workPeriod: time.Duration(workPeriodSec) * time.Second,
		log:        log,
	}

	go w.RunDeletion()

	return w
}

func (w *DeletionWorker) Delete(ctx context.Context, userID string, urls []string) error {
	go func() {
		for _, u := range urls {
			w.entriesCh <- storage.EntryToDelete{ShortURL: u, UserID: userID}
		}
	}()

	return nil
}

func (w *DeletionWorker) RunDeletion() {
	ticker := time.NewTicker(w.workPeriod)

	var toDelete []storage.EntryToDelete

	saveAndReset := func() {
		err := w.storage.MarkDeleted(context.TODO(), toDelete...)
		if err != nil {
			w.log.Errorf("failed to delete urls: %v", err)
			return
		}

		toDelete = nil
	}

	for {
		select {
		case u := <-w.entriesCh:
			if len(toDelete) < cap(w.entriesCh) {
				toDelete = append(toDelete, u)
			} else {
				saveAndReset()
			}
		case <-ticker.C:
			if len(toDelete) == 0 {
				continue
			}

			saveAndReset()
		}
	}
}
