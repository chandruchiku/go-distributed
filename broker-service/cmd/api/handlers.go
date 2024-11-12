package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
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
		app.log(w, &requestPayload.Log)
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
		app.errorJSON(w, err)
		return
	}

	var responsePayload jsonResponse
	responsePayload.Message = "Logged"
	responsePayload.Error = false

	_ = app.writeJson(w, http.StatusAccepted, responsePayload)
}
