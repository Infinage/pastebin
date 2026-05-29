package handlers

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	paste "github.com/infinage/pastebin/pkg"
)

type Application struct {
	st     *paste.Store
	ctx    context.Context
	assets embed.FS
	stop   context.CancelFunc
}

func NewApplication(assets embed.FS) *Application {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return &Application{st: paste.NewEmptyStore(), ctx: ctx, stop: stop, assets: assets}
}

// Periodically cleanup store and remove stale entries
func (app *Application) Serve(addr string) error {
	go app.startBGJobs(30)

	mux := app.routes()
	server := &http.Server{Addr: addr, Handler: mux}
	var wg sync.WaitGroup

	// Interrupt is caught by App.ctx, we need to listen for 
	// int and call shutdown on the server manuallly
	wg.Go(func() {
		<-app.ctx.Done() // block until fired

		// Shutdown immediately stops accepting new cons and drops stale ones
		// Grace duration is for the existing ones - it doesn't close them just returns
		grace := 3 * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
		defer cancel()

		server.Shutdown(shutdownCtx)
	})

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	wg.Wait()
	return nil
}

func (app *Application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HandleHome)
	mux.HandleFunc("GET /paste", app.HandleNewForm)
	mux.HandleFunc("GET /paste/{id}", app.HandleGet)
	mux.HandleFunc("POST /paste", app.HandleInsert)
	mux.HandleFunc("PUT /paste/{id}", app.HandleUpdate)
	mux.HandleFunc("DELETE /paste/{id}", app.HandleDelete)
	mux.Handle("GET /assets/", http.FileServer(http.FS(app.assets)))
	return mux
}

// Periodically cleans up the store of stale entries (every x minutes)
func (app *Application) startBGJobs(minutes int) {
	ticker := time.NewTicker(time.Minute * time.Duration(minutes))
	defer ticker.Stop()

	for {
		select {
		case <-app.ctx.Done():
			return
		case <-ticker.C:
			app.st.Cleanup()
		}
	}
}
