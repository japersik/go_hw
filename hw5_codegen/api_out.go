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

	if login == "" {
		responseWrite(w, r, ErrorAns{"login must be not empty"}, http.StatusBadRequest)
		return
	}

	responseWrite(w, r, SomeAns{Ans: profileparams}, http.StatusOK)
}

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//checkAuth
	if r.Header.Get("X-Auth") != "100500" {
		responseWrite(w, r, ErrorAns{"unauthorized"}, http.StatusForbidden)
		return
	}
	r.ParseForm()
	createparams := CreateParams{}

	//Login --> login
	login := r.Form.Get("login")
	createparams.Login = login

	if login == "" {
		responseWrite(w, r, ErrorAns{"login must be not empty"}, http.StatusBadRequest)
		return
	}

	if len(login) < 10 {
		responseWrite(w, r, ErrorAns{"login len must be >= 10"}, http.StatusBadRequest)
		return
	}

	//Name --> full_name
	name := r.Form.Get("full_name")
	createparams.Name = name

	//Status --> status
	status := r.Form.Get("status")
	createparams.Status = status
	if status == "" {
		status = "user"
	}

	if !(status == "user" || status == "moderator" || status == "admin" || false) {
		responseWrite(w, r, ErrorAns{"status must be one of [user, moderator, admin]"}, http.StatusBadRequest)
		return
	}
	//Age --> age
	age, err := strconv.Atoi(r.Form.Get("age"))
	if err != nil {
		responseWrite(w, r, ErrorAns{"age must be int"}, http.StatusBadRequest)
		return
	}
	createparams.Age = age

	if age > 128 {
		responseWrite(w, r, ErrorAns{"age must be <= 128"}, http.StatusBadRequest)
		return
	}
	if age < 0 {
		responseWrite(w, r, ErrorAns{"age must be >= 0"}, http.StatusBadRequest)
		return
	}

	responseWrite(w, r, SomeAns{Ans: createparams}, http.StatusOK)
}

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//checkAuth
	if r.Header.Get("X-Auth") != "100500" {
		responseWrite(w, r, ErrorAns{"unauthorized"}, http.StatusForbidden)
		return
	}
	r.ParseForm()
	othercreateparams := OtherCreateParams{}

	//Username --> username
	username := r.Form.Get("username")
	othercreateparams.Username = username

	if username == "" {
		responseWrite(w, r, ErrorAns{"username must be not empty"}, http.StatusBadRequest)
		return
	}

	if len(username) < 3 {
		responseWrite(w, r, ErrorAns{"username len must be >= 3"}, http.StatusBadRequest)
		return
	}

	//Name --> account_name
	name := r.Form.Get("account_name")
	othercreateparams.Name = name

	//Class --> class
	class := r.Form.Get("class")
	othercreateparams.Class = class
	if class == "" {
		class = "warrior"
	}

	if !(class == "warrior" || class == "sorcerer" || class == "rouge" || false) {
		responseWrite(w, r, ErrorAns{"class must be one of [warrior, sorcerer, rouge]"}, http.StatusBadRequest)
		return
	}
	//Level --> level
	level, err := strconv.Atoi(r.Form.Get("level"))
	if err != nil {
		responseWrite(w, r, ErrorAns{"level must be int"}, http.StatusBadRequest)
		return
	}
	othercreateparams.Level = level

	if level > 50 {
		responseWrite(w, r, ErrorAns{"level must be <= 50"}, http.StatusBadRequest)
		return
	}
	if level < 1 {
		responseWrite(w, r, ErrorAns{"level must be >= 1"}, http.StatusBadRequest)
		return
	}

	responseWrite(w, r, SomeAns{Ans: othercreateparams}, http.StatusOK)
}

func responseWrite(w http.ResponseWriter, r *http.Request, obj interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(obj)
	w.Write(b)
}
