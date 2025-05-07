package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

func (a *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	a.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *application) notFound(w http.ResponseWriter) {
	a.clientError(w, http.StatusNotFound)
}

func (a application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := a.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		a.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		a.serverError(w, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (a *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()

	r.PostForm.Del("gorilla.csrf.Token")
	if err != nil {
		return err
	}

	err = a.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		if dst == nil {
			panic(err)
		}

		a.errorLog.Println(err)
		return err
	}
	return nil
}

func (a application) newTemplateData(w http.ResponseWriter, r *http.Request) *templateData {

	session := r.Context().Value(sessionContextKey).(*sessions.Session)
	var flashMsg string
	flash := session.Flashes()
	if len(flash) != 0 {
		flashMsg = flash[0].(string)
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	session.Save(r, w)

	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           flashMsg,
		IsAuthenticated: a.isAuthenticated(r),
		CSRFField:       csrf.TemplateField(r),
	}
}

func (a *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
