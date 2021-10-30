package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

//type dataRow struct {
//	Id int `xml:"id"`
//	Guid string `xml:"guid"`
//	IsActive bool `xml:"isActive"`
//	Balance string `xml:"balance"`
//	Picture string `xml:"picture"`
//	Age int `xml:"age"`
//	EyeColor string `xml:"eyeColor"`
//	FirstName string `xml:"first_name"`
//	LastName string `xml:"last_name"`
//	Gender string `xml:"gender"`
//	Company string `xml:"company"`
//	Email string `xml:"email"`
//	Phone string `xml:"phone"`
//	Address string `xml:"address"`
//	About string `xml:"about"`
//}
const pageSize = 25
type dataRow struct {
	Id int `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName string `xml:"last_name"`
	Age int `xml:"age"`
	About string `xml:"about"`
	Gender string `xml:"gender"`
}
type DataTable struct {
	Row []dataRow `xml:"row"`
}

func readUsers() []User{
	file,err := os.Open("dataset.xml")
	if err!= nil{
		panic(err)
	}
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	io.Copy(buf,file)

	dataSet := &DataTable{}
	xml.Unmarshal(buf.Bytes(),dataSet)

	var users []User

	for _, user := range dataSet.Row {
		users = append(users, User{
			Id: user.Id,
			Name: user.FirstName + user.LastName,
			Age: user.Age,
			About: user.About,
			Gender: user.Gender,
		})
	}
	return users
}
func SearchSuccess(w http.ResponseWriter, r *http.Request)  {
	users := readUsers()

	offset, _ := strconv.Atoi(r.FormValue("offset"))
	limit, _ := strconv.Atoi(r.FormValue("limit"))

	endRow := offset + limit
	users = users[ offset: endRow ]

	jsonResponse, err := json.Marshal(users)

	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
func SearchIgnoreLimit(w http.ResponseWriter, r *http.Request)  {
	jsonResponse, err := json.Marshal(readUsers())

	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func TestGoodResponse(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchSuccess))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}

	searchRequest := SearchRequest{
		Limit: 26,
		Offset: 0,
	}

	_, err := searchClient.FindUsers(searchRequest)

	if err != nil {
		t.Error("Error on success request")
	}

	searchRequest.Limit = -1
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "limit must be > 0" {
		t.Error("Limit error: not 'limit must be > 0'")
	}

	searchRequest.Limit = 1
	searchRequest.Offset = -1
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "offset must be > 0" {
		t.Error("Offset error: not 'offset must be > 0'")
	}
}

func TestIgnoreLimit(t *testing.T) {
	limit := 10
	ts := httptest.NewServer(http.HandlerFunc(SearchIgnoreLimit))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}

	searchRequest := SearchRequest{Limit: limit}

	result, err := searchClient.FindUsers(searchRequest)

	if err != nil {
		t.Error("Error on success request")
	}
	if len(result.Users)==limit{
		t.Error(fmt.Sprintf("IgnoreLimit test failed: %v(len(result.Users))  == %v(limit)",len(result.Users),limit))
	}
}

func TestBadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		io.WriteString(writer, `I'm not JSON'`)
	}))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil || !strings.Contains(err.Error(),"cant unpack result json:") {
		t.Error("BadJSON test failed")
	}
}

func TestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second * 2)
		writer.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil || !strings.Contains(err.Error(),"timeout for") {
		t.Error("BadJSON test failed")
	}
}

func TestBadAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil || !strings.Contains(err.Error(),"Bad AccessToken") {
		t.Error("Bad AccessToken test failed")
	}
}

func TestUnknownError(t *testing.T) {
	searchClient := &SearchClient{"","Err URL"}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil || !strings.Contains(err.Error(),"unknown error") {
		t.Error("unknown error test failed")
	}
}

func TestStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	searchClient := &SearchClient{"",ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil || !strings.Contains(err.Error(),"SearchServer fatal error") {
		t.Error("StatusInternalServerError test failed")
	}
}
