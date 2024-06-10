package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Broker service is up and running",
	}

	if err := app.writeJSON(w, http.StatusOK, payload); err != nil {
		app.errorJSON(w, err)
	}
}

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

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, r, requestPayload.Auth)
	case "log":
		app.Log(w, r, requestPayload.Log)
	default:
		app.errorJSON(w, errors.New("unknown action"), http.StatusBadRequest)
	}
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request, auth AuthPayload) {
	jsonData, _ := json.MarshalIndent(auth, "", "\t")

	response, err := http.Post("http://authentication-service/authenticate", "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusOK {
		app.errorJSON(w, errors.New("unexpected error"), response.StatusCode)
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: jsonFromService.Message,
		Data:    jsonFromService.Data,
	}

	if err := app.writeJSON(w, http.StatusOK, payload); err != nil {
		app.errorJSON(w, err)
	}
}

func (app *Config) Log(w http.ResponseWriter, r *http.Request, log LogPayload) {
	jsonData, _ := json.MarshalIndent(log, "", "\t")

	response, err := http.Post("http://logger-service/log", "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Log submitted",
	}

	if err := app.writeJSON(w, http.StatusOK, payload); err != nil {
		app.errorJSON(w, err)
	}
}
