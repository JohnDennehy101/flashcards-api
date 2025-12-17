package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/flashcards", app.listFlashcardsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/flashcards", app.createFlashcardHandler)
	router.HandlerFunc(http.MethodGet, "/v1/flashcards/:id", app.showFlashcardHandler)
	router.HandlerFunc(http.MethodPut, "/v1/flashcards/:id", app.updateFlashcardHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/flashcards/:id", app.deleteFlashcardHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
