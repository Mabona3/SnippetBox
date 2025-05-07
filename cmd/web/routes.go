package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"snippetbox.mabona3.net/ui"
)

func (a *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))

	protected := alice.New(a.requireAuthentication)
	authing := alice.New(a.requireNoAuthentication)

	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/", a.home)
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", a.snippetView)

	router.Handler(http.MethodGet, "/user/signup", authing.ThenFunc(a.userSignup))
	router.Handler(http.MethodPost, "/user/signup", authing.ThenFunc(a.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", authing.ThenFunc(a.userLogin))
	router.Handler(http.MethodPost, "/user/login", authing.ThenFunc(a.userLoginPost))

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(a.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(a.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(a.userLogoutPost))

	return alice.New(
		a.recoverPanic,
		a.logRequest,
		secureHeaders,
		a.authenticate,
		a.noSurf,
		a.InitializeSession,
	).Then(router)
}
