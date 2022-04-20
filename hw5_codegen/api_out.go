package main

import (
	"encoding/json"
	"net/http"
)

func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		api.handlerProfile(w, r)
	case "/user/create":
		api.handlerCreate(w, r)
	default:
		responseWrite(w, r, "some error", http.StatusBadRequest)
	}
}

func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		api.handlerCreate(w, r)
	default:
		responseWrite(w, r, "some error", http.StatusBadRequest)
	}
}

func (api *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	//write Body
}

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//write Body
}

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//write Body
}

func responseWrite(w http.ResponseWriter, r *http.Request, obj interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(obj)
	w.Write(b)
}
