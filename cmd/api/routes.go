package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/flashcards", app.requirePermission("flashcards:read", app.listFlashcardsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/flashcards", app.requirePermission("flashcards:write", app.createFlashcardHandler))
	router.HandlerFunc(http.MethodGet, "/v1/flashcards/:id", app.requirePermission("flashcards:read", app.showFlashcardHandler))
	router.HandlerFunc(http.MethodPut, "/v1/flashcards/:id", app.requirePermission("flashcards:write", app.updateFlashcardHandler))
	router.HandlerFunc(http.MethodPost, "/v1/flashcards/:id/review", app.requirePermission("flashcards:write", app.reviewFlashcardHandler))
	router.HandlerFunc(http.MethodPost, "/v1/flashcards/:id/reset", app.requirePermission("flashcards:write", app.resetFlashcardHandler))

	router.HandlerFunc(http.MethodDelete, "/v1/flashcards/:id", app.requirePermission("flashcards:write", app.deleteFlashcardHandler))

	router.HandlerFunc(http.MethodGet, "/v1/stats/flashcards", app.requirePermission("flashcards:read", app.showFlashcardStatsHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
