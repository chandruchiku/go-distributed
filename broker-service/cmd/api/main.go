package main

import (
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const httpPort = "8080"

type Config struct {
	// HTTPPort string `envconfig:"HTTP_PORT" default:"80"`
	Rabbit *amqp.Connection
}

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connectToRabbitMQ()

	// if connection fails, log the error and exit
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := &Config{
		Rabbit: rabbitConn,
	}
	log.Println("Starting server on port " + httpPort)

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	// connect to rabbitmq
	var counts int64
	var backOff = 1 * time.Second
	var conn *amqp.Connection

	// try to connect to rabbitmq
	for {
		c, err := amqp.Dial("amqp://rabbitmq:password@rabbitmq")
		if err != nil {
			log.Println("Failed to connect to RabbitMQ. Retrying in", backOff)
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			conn = c
			break
		}

		if counts > 5 {
			log.Println("Failed to connect to RabbitMQ after 5 retries")
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Retrying in", backOff)
		time.Sleep(backOff)
		continue
	}

	// return the connection
	return conn, nil
}
