package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const httpPort = 8080

var retryCounts = 10

type Config struct {
	DB   *sql.DB
	Repo data.Repository
}

func main() {
	log.Println("Starting Authentication server on port", httpPort)

	// connect database
	conn := connectToDB()
	if conn == nil {
		log.Fatal("Failed to connect to database")
	}

	// Setup config
	app := &Config{
		DB:   conn,
		Repo: data.New(conn),
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	log.Println("Authentication server is running on port", httpPort)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres Database is not ready yet...")
			log.Println("Retrying in 2 seconds...")
			retryCounts--
			time.Sleep(2 * time.Second)
		} else {
			log.Println("Postgres Database is connected!")
			return connection
		}

		if retryCounts == 0 {
			log.Fatal("Max retries exceeded. Exiting...")
			return nil
		}
	}
}
