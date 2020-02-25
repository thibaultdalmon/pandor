package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var johnDoe = UserModel{
	Name: "John Doe",
	Mail: "john@doe.fr",
}

var johnjohnDoe = UserModel{
	Name: "John John Doe",
	Mail: "john@doe.fr",
}

func request(method, url string, body interface{}) *httptest.ResponseRecorder {
	e := newRouter()
	j, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	return res
}

func response(res *httptest.ResponseRecorder, v interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func TestGetMessage(t *testing.T) {
	// Make the request
	res := request("GET", "/?message=Salut", nil)
	assert.Equal(t, 200, res.Code)

	// Check content field
	var answer Message
	response(res, &answer)
	assert.Equal(t, "Salut", answer.Content)
}

func TestAPI(t *testing.T) {
	res := request("PUT", "/?mail="+johnDoe.Mail+"&name="+johnDoe.Name, nil)
	assert.Equal(t, 200, res.Code)

	res = request("GET", "/?mail="+johnDoe.Mail, nil)
	assert.Equal(t, johnDoe.Name, answer.Content)

	res = request("POST", "/?mail="+johnDoe.Mail+"&name="+johnjohnDoe.Name, nil)
	assert.Equal(t, 200, res.Code)

	res = request("GET", "/?mail="+johnDoe.Mail, nil)
	assert.Equal(t, johnjohnDoe.Name, answer.Content)

	res = request("DELETE", "/?mail="+johnDoe.Mail, nil)
	assert.Equal(t, 200, res.Code)

	res = request("GET", "/?mail="+johnDoe.Mail, nil)
	assert.Equal(t, 401, res.Code)
}
