package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"fire-go/graph"
	"fire-go/internal/db"
	"fire-go/internal/handlers"
	"fire-go/internal/logger"
	"fire-go/internal/middleware"
)

func main() {

	db.Connect()
	logger.Init()

	logger.Log.Info("server_starting", "port", 8081)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver()}))

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)

	// Rotas REST
	r.Get("/fire_intersect", handlers.GetFireIntersectFiltered)

	// Rotas GraphQL
	r.Handle("/graphql", srv)
	r.Handle("/", playground.Handler("GraphQL playground", "/graphql"))

	log.Println("API rodando em :8081")
	log.Println("GraphQL playground disponível em http://localhost:8081/")
	http.ListenAndServe(":8081", r)
}
