package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"

	"boilerplate/graph"
)

// SetupRoutes registers the GraphQL and Playground endpoints.
func SetupRoutes(app *fiber.App) {
	// Initialize the GraphQL server with the generated executable schema.
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	// Mount the GraphQL endpoint using Fiber's http adaptor.
	app.Post("/query", adaptor.HTTPHandler(srv))

	// Optionally provide a GraphQL playground GET endpoint.
	playSrv := playground.Handler("GraphQL playground", "/query")
	app.Get("/", adaptor.HTTPHandler(playSrv))
}
