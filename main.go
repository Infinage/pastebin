package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/infinage/pastebin/internal/views"
	paste "github.com/infinage/pastebin/pkg"
)

//go:embed assets/*
var assets embed.FS

func main() {
	mux := http.NewServeMux()
	st := paste.NewEmptyStore()

	mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		pastes := st.ListPublic()
		if r.Header.Get("Hx-Request") == "true" {
			views.List(pastes).Render(context.Background(), w)
			views.AddEntryBtn().Render(context.Background(), w)
			return
		}

		list := templ.ComponentFunc(func(ctx context.Context, w1 io.Writer) error {
			err := views.List(pastes).Render(context.Background(), w1)
			if err != nil {
				return err
			}
			return views.AddEntryBtn().Render(context.Background(), w1)
		})

		ctx := templ.WithChildren(context.Background(), list)
		views.Layout().Render(ctx, w)
	})

	mux.HandleFunc("GET /paste/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		model, ok := st.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}

		// If htmx request, simply return component that needs rendering
		if r.Header.Get("Hx-Request") == "true" {
			views.Form(model, false).Render(context.Background(), w)
			return
		}

		form := templ.ComponentFunc(func(ctx context.Context, w1 io.Writer) error {
			return views.Form(model, false).Render(context.Background(), w1)
		})

		ctx := templ.WithChildren(context.Background(), form)
		views.Layout().Render(ctx, w)
	})

	mux.HandleFunc("GET /paste", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Hx-Request") == "true" {
			views.Form(paste.Model{}, true).Render(context.Background(), w)
			return
		}

		form := templ.ComponentFunc(func(ctx context.Context, w1 io.Writer) error {
			return views.Form(paste.Model{}, true).Render(context.Background(), w1)
		})

		ctx := templ.WithChildren(context.Background(), form)
		views.Layout().Render(ctx, w)
	})

	mux.HandleFunc("DELETE /paste/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		st.Delete(id)
		w.Header().Set("HX-Redirect", "/")
	})

	mux.HandleFunc("POST /paste", func(w http.ResponseWriter, r *http.Request) {
		fields, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := st.Insert(fields.content, fields.expiry, fields.visibility)
		w.Header().Set("HX-Redirect", "/paste/"+id)
	})

	mux.HandleFunc("PUT /paste/{id}", func(w http.ResponseWriter, r *http.Request) {
		fields, err := parseForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		st.Delete(r.PathValue("id")) // Delete old entry
		id := st.Insert(fields.content, fields.expiry, fields.visibility)
		w.Header().Set("HX-Redirect", "/paste/"+id)
	})

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", mux)
}

type modelFormFields struct {
	content    string
	expiry     time.Duration
	visibility paste.Visibility
}

func parseForm(r *http.Request) (modelFormFields, error) {
	if err := r.ParseForm(); err != nil {
		return modelFormFields{}, fmt.Errorf("Error parsing form")
	}

	visibilityRaw := r.FormValue("visibility")
	visibility, err := strconv.Atoi(visibilityRaw)
	if err != nil || visibility != 1 && visibility != 2 {
		return modelFormFields{}, fmt.Errorf("Visibility field must be 1 (public) or 2 (unlisted)")
	}

	expiryMinutesRaw := r.FormValue("expiry")
	expiryMinutes, err := strconv.Atoi(expiryMinutesRaw)
	if err != nil || expiryMinutes < 1 || expiryMinutes > 110376000 {
		return modelFormFields{}, fmt.Errorf("Expiry field must be between [1, 110376000]")
	}

	content := r.FormValue("content")
	expiry := time.Minute * time.Duration(expiryMinutes)
	res := modelFormFields{content: content, expiry: expiry, visibility: paste.Visibility(visibility)}
	return res, nil
}
