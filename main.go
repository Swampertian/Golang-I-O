package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"fire-go/internal/db"
	"fire-go/internal/handlers"
)

func main() {
	db.Connect()

	r := chi.NewRouter()
	r.Get("/fire_intersect", handlers.GetFireIntersectFiltered)

	log.Println("API rodando em :8081")
	http.ListenAndServe(":8081", r)
}
