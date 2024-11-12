package main

import (
	"log-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// read the json
	var requestPayload JSONPayload
	_ = app.readJson(w, r, &requestPayload)

	// insert the log
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// return the response
	resp := jsonResponse{
		Error:   false,
		Message: "Log entry created",
	}

	app.writeJson(w, http.StatusAccepted, resp)
}
