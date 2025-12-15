package graph

import (
	"fire-go/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DB *pgxpool.Pool
}

func NewResolver() *Resolver {
	return &Resolver{
		DB: db.Pool,
	}
}
