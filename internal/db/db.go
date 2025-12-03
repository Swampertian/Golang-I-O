package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não definido")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Erro ParseConfig:", err)
	}

	cfg.MaxConns = 15
	cfg.MinConns = 2
	cfg.MaxConnIdleTime = 5 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	log.Println("Conectado ao PostgreSQL")
}
