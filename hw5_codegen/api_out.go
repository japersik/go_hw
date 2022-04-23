package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorAns struct {
	Err string `json:"error"`
}

type SomeAns struct {
	ErrorAns
	Ans interface{} `json:"response"`
}

func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		api.handlerProfile(w, r)
	case "/user/create":
		api.handlerCreate(w, r)
	default:
		responseWrite(w, r, ErrorAns{"unknown method"}, http.StatusNotFound)
	}
}

func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		api.handlerCreate(w, r)
	default:
		responseWrite(w, r, ErrorAns{"unknown method"}, http.StatusNotFound)
	}
}

func (api *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	profileparams := ProfileParams{}

	//Login --> login
	login := r.Form.Get("login")
	profileparams.Login = login

	responseWrite(w, r, SomeAns{Ans: profileparams}, http.StatusOK)
}

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//checkAuth
	if r.Header.Get("X-Auth") != "100500" {
		responseWrite(w, r, ErrorAns{"unauthorized"}, http.StatusForbidden)
	}
	r.ParseForm()
	createparams := CreateParams{}

	//Login --> login
	login := r.Form.Get("login")
	createparams.Login = login

	//Name --> full_name
	name := r.Form.Get("full_name")
	createparams.Name = name

	//Status --> status
	status := r.Form.Get("status")
	createparams.Status = status

	//Age --> age
	age, err := strconv.Atoi(r.Form.Get("age"))
	if err != nil {
		responseWrite(w, r, ErrorAns{"age must be int"}, http.StatusBadRequest)
		return
	}

	createparams.Age = age

	responseWrite(w, r, SomeAns{Ans: createparams}, http.StatusOK)
}

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//checkAuth
	if r.Header.Get("X-Auth") != "100500" {
		responseWrite(w, r, ErrorAns{"unauthorized"}, http.StatusForbidden)
	}
	r.ParseForm()
	othercreateparams := OtherCreateParams{}

	//Username --> username
	username := r.Form.Get("username")
	othercreateparams.Username = username

	//Name --> account_name
	name := r.Form.Get("account_name")
	othercreateparams.Name = name

	//Class --> class
	class := r.Form.Get("class")
	othercreateparams.Class = class

	//Level --> level
	level, err := strconv.Atoi(r.Form.Get("level"))
	if err != nil {
		responseWrite(w, r, ErrorAns{"level must be int"}, http.StatusBadRequest)
		return
	}

	othercreateparams.Level = level

	responseWrite(w, r, SomeAns{Ans: othercreateparams}, http.StatusOK)
}

func responseWrite(w http.ResponseWriter, r *http.Request, obj interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(obj)
	w.Write(b)
}
