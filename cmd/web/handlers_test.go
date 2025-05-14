package main

import (
	"net/http"
	"net/url"
	"testing"

	"snippetbox.mabona3.net/internal/assert"
)

func TestPing(t *testing.T) {
	a := newTestApplication(t)

	ts := newTestServer(t, a.routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/ping")
	if err != nil{
		t.Fatal(err)
	}
	defer rs.Body.Close()
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestSnippetView(t *testing.T) {
	a := newTestApplication(t)
	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name string
		urlPath string
		wantCode int
		wantBody string
	} {
		{
			name: "Valid ID",
			urlPath: "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name: "Non-existent ID",
			urlPath: "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name: "Negative ID",
			urlPath: "/snippet/view/-1",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "Decimal ID",
			urlPath: "/snippet/view/1.23",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "String ID",
			urlPath: "/snippet/view/foo",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "Empty ID",
			urlPath: "/snippet/view",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignup(t *testing.T) {
	a := newTestApplication(t)
	ts := newTestServer(t, a.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)
	const (
		validName = "Bob"
		validPassword = "validPa$$word"
		validEmail = "bob@example.com"
		formTag = "<form action='/user/signup' method='post' novalidate>"
	)

	tests := []struct {
		name string
		userName string
		userEmail string
		userPassword string
		csrfToken string
		wantCode int
		wantFormTag string
	} {
		{
			name: "Valid submission",
			userName: validName,
			userEmail: validEmail,
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusSeeOther,
		},
		{
			name: "Invalid CSRF Token",
			userName: validName,
			userEmail: validEmail,
			userPassword: validPassword,
			csrfToken: "wrongToken",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "Empty name",
			userName: "",
			userEmail: validEmail,
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
		{
			name: "Empty email",
			userName: validName,
			userEmail: "",
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
		{
			name: "Empty password",
			userName: validName,
			userEmail: validEmail,
			userPassword: "",
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
		{
			name: "Invalid email",
			userName: validName,
			userEmail: "bob@example.",
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
		{
			name: "Short password",
			userName: validName,
			userEmail: "pa$$",
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
		{
			name: "Duplicate email",
			userName: validName,
			userEmail: "dupe@example.com",
			userPassword: validPassword,
			csrfToken: validCSRFToken,
			wantCode: http.StatusBadRequest,
			wantFormTag: formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)
			
			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}

}
