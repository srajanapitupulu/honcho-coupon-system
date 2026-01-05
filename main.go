package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	coupon "github.com-personal/srajanapitupulu/honcho-coupon-system/pkg/coupon"
)

var db *sql.DB

func main() {
	// Database Connection
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/honcho_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Router Setup using Gin
	r := coupon.SetupRouter(db)

	// Start Server
	log.Println("Server starting on :8080")
	r.Run(":8080")
}
