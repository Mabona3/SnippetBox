package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (a application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (a *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				a.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *application) InitializeSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := a.Store.Get(r, "session")
		session.Options.SameSite = http.SameSiteLaxMode
		ctx := context.WithValue(r.Context(), "session", session)
		r = r.WithContext(ctx)

		if err != nil {
			a.serverError(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *application)customCSRFErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CSRF validation failed")
	http.Error(w, "CSRF token invalid or expired", http.StatusBadRequest)
}

func (a *application) noSurf(next http.Handler) http.Handler {
	csrfHandler := csrf.Protect([]byte(
		os.Getenv("SECRET_KEY")),
		csrf.Secure(true),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteDefaultMode),
		csrf.ErrorHandler(http.HandlerFunc(a.customCSRFErrorHandler)),
	)
	return csrfHandler(next)
}

func (a *application) requireNoAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.isAuthenticated(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
