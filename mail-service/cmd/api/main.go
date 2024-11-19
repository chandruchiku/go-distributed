package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer *Mail
}

const httpPort = 8080

func main() {
	app := &Config{}

	app.Mailer = createMail()

	log.Println("Starting mail-service on port", httpPort)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func createMail() *Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	return &Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
	}
}
