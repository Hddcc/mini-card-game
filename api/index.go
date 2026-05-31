package handler

import (
	"net/http"
	"sync"

	"mini-card-game/internal/server"
)

var (
	once    sync.Once
	app     http.Handler
	initErr error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		var cleanup func()
		app, _, cleanup, initErr = server.New(server.Options{
			AllowMissingDatabase: true,
		})
		_ = cleanup
	})

	if initErr != nil {
		http.Error(w, initErr.Error(), http.StatusInternalServerError)
		return
	}

	app.ServeHTTP(w, r)
}
