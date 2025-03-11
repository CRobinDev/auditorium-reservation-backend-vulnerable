package database

import (
	"fmt"
	"github.com/nathakusuma/auditorium-reservation-backend/pkg/log"
	"sync"
	"time"

	// pgx driver for postgres
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
)

var (
	db   *sqlx.DB
	once sync.Once
)

func NewPostgresPool(host, port, user, pass, dbName string) *sqlx.DB {
	once.Do(func() {
		dataSourceName := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, pass, dbName,
		)

		pool, err := sqlx.Connect("pgx", dataSourceName)
		if err != nil {
			log.Fatal(map[string]interface{}{
				"error": err.Error(),
			}, "[DB][NewPostgresPool] failed to connect to database")
		}

		pool.SetMaxOpenConns(100)
		pool.SetMaxIdleConns(10)
		pool.SetConnMaxLifetime(60 * time.Minute)

		db = pool
	})

	return db
}
