package main

import (
	"errors"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		EmailId  string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.readJson(w, r, &payload); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Repo.User.GetByEmail(payload.EmailId)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(payload.Password)
	if !valid || err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: "Successfully authenticated",
		Data:    user,
	}

	app.writeJson(w, http.StatusAccepted, response)
}
