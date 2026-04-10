package db

import (
	"log"

	"github.com/pocketbase/pocketbase"
)

var app *pocketbase.PocketBase

// Init initializes the PocketBase instance but does not start its server yet.
func Init() *pocketbase.PocketBase {
	app = pocketbase.New()
	return app
}

// GetApp returns the initialized PocketBase core app instance.
func GetApp() *pocketbase.PocketBase {
	if app == nil {
		log.Fatal("PocketBase is not initialized")
	}
	return app
}
