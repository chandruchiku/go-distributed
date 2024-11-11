package main

import (
	"log"
	"net/http"
)

const httpPort = "8080"

type Config struct {
	HTTPPort string `envconfig:"HTTP_PORT" default:"80"`
}

func main() {
	app := &Config{}
	log.Println("Starting server on port " + httpPort)

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
