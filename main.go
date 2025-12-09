package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"fire-go/internal/db"
	"fire-go/internal/handlers"
	"fire-go/internal/logger"
	"fire-go/internal/middleware"
)

func main() {

	db.Connect()
	logger.Init()
	logger.Log.Info("server_starting", "port", 8081)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Get("/fire_intersect", handlers.GetFireIntersectFiltered)

	log.Println("API rodando em :8081")
	http.ListenAndServe(":8081", r)
}
