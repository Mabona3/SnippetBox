package main

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
const sessionContextKey = contextKey("session")
