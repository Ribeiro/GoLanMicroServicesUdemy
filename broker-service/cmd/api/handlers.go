package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
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

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)

	default:
		app.errorJSON(w, errors.New("Unknown Action"), http.StatusBadRequest)
	}

}

func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {
	//build json to send to Authentication Microservice
	jsonData, _ := json.MarshalIndent(authPayload, "", "\t")

	//call Authentication Microservice
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
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

	//make sure send back correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusUnauthorized)
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("Error calling Authentication Service"), http.StatusUnauthorized)
		return
	}

	//create variable to read response.Body into
	var jsonFromService jsonResponse

	//decode json received from Auth Service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJson(w, http.StatusAccepted, payload)

}
