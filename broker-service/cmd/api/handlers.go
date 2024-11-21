package main

import (
	"broker-service/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := &jsonResponse{
		Error:   false,
		Message: "Hit the broker endpoint",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	log.Printf("Received request: %s", requestPayload.Action)

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, &requestPayload.Auth)
	case "log":
		// app.log(w, &requestPayload.Log)
		// app.logEventViaRabbit(w, &requestPayload.Log)
		app.logEventViaRPC(w, &requestPayload.Log)
	case "mail":
		app.mail(w, &requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, payload *AuthPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")
	log.Printf("Authenticating: %s", jsonData)

	request, err := http.NewRequest("POST", "http://authentication-service:8080/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error authenticating"))
		return
	}

	// Read response body
	var authResponse jsonResponse
	// Decode the response body into the authResponse struct
	err = json.NewDecoder(response.Body).Decode(&authResponse)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if authResponse.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	// Set the auth token in the response
	var responsePayload jsonResponse
	responsePayload.Message = "Authenticated"
	responsePayload.Error = false
	responsePayload.Data = authResponse.Data

	_ = app.writeJson(w, http.StatusOK, responsePayload)
}

func (app *Config) log(w http.ResponseWriter, payload *LogPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")
	log.Printf("Logging: %s", jsonData)

	request, err := http.NewRequest("POST", "http://logger-service:8080/log", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error logging"))
		return
	}

	var responsePayload jsonResponse
	responsePayload.Message = "Logged"
	responsePayload.Error = false

	_ = app.writeJson(w, http.StatusAccepted, responsePayload)
}

func (app *Config) mail(w http.ResponseWriter, payload *MailPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")
	log.Printf("Mailing: %s", jsonData)

	request, err := http.NewRequest("POST", "http://mailer-service:8080/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	var responsePayload jsonResponse
	responsePayload.Message = "Mailed"
	responsePayload.Error = false

	_ = app.writeJson(w, http.StatusAccepted, responsePayload)
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logEventViaRPC(w http.ResponseWriter, l *LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", payload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var resp jsonResponse
	resp.Message = "Logged via RPC"
	resp.Error = false

	_ = app.writeJson(w, http.StatusAccepted, resp)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l *LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var resp jsonResponse
	resp.Message = "Logged via RabbitMQ"
	resp.Error = false

	_ = app.writeJson(w, http.StatusAccepted, resp)
}

func (app *Config) pushToQueue(name, msg string) error {
	emmitter, err := event.NewEventEmmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(payload, "", "\t")
	err = emmitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
