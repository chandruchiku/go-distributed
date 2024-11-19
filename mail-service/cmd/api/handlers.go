package main

import (
	"log"
	"net/http"
)

type mailMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	var requestPayload mailMessage

	log.Println("Received request to send email")

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      []string{requestPayload.To},
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Println("Error sending email:", err)
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Email sent " + requestPayload.To,
	}

	log.Println("Email sent to", requestPayload.To)

	app.writeJson(w, http.StatusAccepted, payload)
}
