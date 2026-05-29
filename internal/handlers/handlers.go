package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/infinage/pastebin/internal/views"
	paste "github.com/infinage/pastebin/pkg"
)

func (app *Application) HandleHome(w http.ResponseWriter, r *http.Request) {
	pastes := app.st.ListPublic()
	if r.Header.Get("Hx-Request") == "true" {
		views.List(pastes).Render(app.ctx, w)
		views.AddEntryBtn().Render(app.ctx, w)
		return
	}

	list := templ.ComponentFunc(func(ctx context.Context, cw io.Writer) error {
		err := views.List(pastes).Render(app.ctx, cw)
		if err != nil {
			return err
		}
		return views.AddEntryBtn().Render(app.ctx, cw)
	})

	ctx := templ.WithChildren(app.ctx, list)
	views.Layout().Render(ctx, w)
}

func (app *Application) HandleGet(w http.ResponseWriter, r *http.Request) {
	model, ok := app.st.Get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	// If htmx request, simply return component that needs rendering
	if r.Header.Get("Hx-Request") == "true" {
		views.Form(model, false).Render(app.ctx, w)
		return
	}

	// Otherwise build component and embed into layout
	form := templ.ComponentFunc(func(ctx context.Context, w1 io.Writer) error {
		return views.Form(model, false).Render(app.ctx, w1)
	})
	ctx := templ.WithChildren(app.ctx, form)
	views.Layout().Render(ctx, w)
}

func (app *Application) HandleNewForm(w http.ResponseWriter, r *http.Request) {
	// If htmx request just render the form
	if r.Header.Get("Hx-Request") == "true" {
		views.Form(paste.Model{}, true).Render(context.Background(), w)
		return
	}

	// Otherwise build the form and embed into layout
	form := templ.ComponentFunc(func(ctx context.Context, w1 io.Writer) error {
		return views.Form(paste.Model{}, true).Render(context.Background(), w1)
	})
	ctx := templ.WithChildren(context.Background(), form)
	views.Layout().Render(ctx, w)
}

func (app *Application) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	app.st.Delete(id)
	w.Header().Set("HX-Redirect", "/")
}

func (app *Application) HandleInsert(w http.ResponseWriter, r *http.Request) {
	fields, err := parseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := app.st.Insert(fields.content, fields.expiry, fields.visibility)
	w.Header().Set("HX-Redirect", "/paste/"+id)
}

func (app *Application) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	fields, err := parseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.st.Delete(r.PathValue("id")) // Delete old entry
	id := app.st.Insert(fields.content, fields.expiry, fields.visibility)
	w.Header().Set("HX-Redirect", "/paste/"+id)
}
