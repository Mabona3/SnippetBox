package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/sessions"
	"snippetbox.mabona3.net/internal/models"
)

type templateData struct {
	CurrentYear int
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
	Form        any
	Flash       string
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func (a application) newTemplateData(r *http.Request) *templateData {

	session := r.Context().Value("session").(*sessions.Session)
	var flashMsg string
	flash := session.Flashes()
	if len(flash) != 0 {
		flashMsg = flash[0].(string)
	}

	return &templateData{
		CurrentYear: time.Now().Year(),
		Flash:       flashMsg,
	}
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
