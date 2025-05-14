package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"snippetbox.mabona3.net/internal/models"
	"snippetbox.mabona3.net/internal/validator"
)

type snippetCreateForm struct {
	Title               string
	Content             string
	Expires             int
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (a *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := a.snippets.Latest()
	if err != nil {
		a.serverError(w, err)
		return
	}

	data := a.newTemplateData(w, r)
	data.Snippets = snippets

	a.render(w, http.StatusOK, "home.html", data)
}

func (a *application) snippetView(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionContextKey).(*sessions.Session)
	if session == nil {
		a.serverError(w, errors.New("No Session Initialized"))
		return
	}

	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	snippet, err := a.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			a.notFound(w)
		} else {
			a.serverError(w, err)
		}
		return
	}

	data := a.newTemplateData(w, r)
	data.Snippet = snippet

	err = session.Save(r, w)
	if err != nil {
		a.serverError(w, err)
		return
	}
	a.render(w, http.StatusOK, "view.html", data)
}

func (a *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(w, r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	a.render(w, http.StatusOK, "create.html", data)
}

func (a *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	session := r.Context().Value(sessionContextKey).(*sessions.Session)
	if session == nil {
		a.serverError(w, models.ErrSessionNotFound)
		return
	}

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PremittedValue(form.Expires, 1, 7, 365), "expires", "This firld must equal 1, 7 or 365")

	if !form.Valid() {
		data := a.newTemplateData(w, r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := a.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		a.serverError(w, err)
		return
	}

	session.AddFlash("Snippet successfully created!")
	err = session.Save(r, w)
	if err != nil {
		a.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (a *application) Neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			a.notFound(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(w, r)
	data.Form = userSignupForm{}
	a.render(w, http.StatusOK, "signup.html", data)
}

func (a *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := a.newTemplateData(w, r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = a.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email Address is already in use")
			data := a.newTemplateData(w, r)
			data.Form = form
			a.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			a.serverError(w, err)
		}

		return
	}

	session := r.Context().Value(sessionContextKey).(*sessions.Session)
	session.AddFlash("Your signup was successful . Please Login.")
	session.Save(r, w)

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (a *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(w, r)
	data.Form = userLoginForm{}
	a.render(w, http.StatusOK, "login.html", data)
}

func (a *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := a.newTemplateData(w, r)
		data.Form = form
		a.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := a.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := a.newTemplateData(w, r)
			data.Form = form
			a.render(w, http.StatusUnprocessableEntity, "login.html", data)
			return
		}
		a.serverError(w, err)
		return
	}

	session, err := a.Store.New(r, "authsession")
	if err != nil {
		a.serverError(w, err)
		return
	}

	session.Options.SameSite = http.SameSiteLaxMode
	session.Values["userId"] = id
	session.Save(r, w)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (a *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionContextKey).(*sessions.Session)

	session.AddFlash("You've logged out successfully!")
	session.Save(r, w)

	authsession, err := a.Store.Get(r, "authsession")
	if err != nil {
		a.serverError(w, err)
		return
	}

	authsession.Options.MaxAge = -1
	authsession.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
