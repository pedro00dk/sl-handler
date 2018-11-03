package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"./database"

	"github.com/orisano/uds"
)

const (
	dockerSocketPath = "/var/run/docker.sock"
)

var (
	dockerAPIClient = uds.NewClient(dockerSocketPath)
	db              = database.Database{}
)

func main() {
	db.Connect()

	http.HandleFunc("/function/", function)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/call", call)
	http.ListenAndServe(":8000", nil)
}

func function(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		//methodGet(res, req)
	case "POST":
		functionPost(res, req)
	case "DELETE":
		functionDelete(res, req)
	}
}

func functionPost(res http.ResponseWriter, req *http.Request) {
	name, memory, code, pack := ExtractFunction(res, req.Body)
	fmt.Println(len(db.SelectFunction(name)))
	if len(db.SelectFunction(name)) == 0 {
		db.InsertFunction(name, memory, code, pack)
		res.Write([]byte(fmt.Sprintf("Function Created [%v] %v\n", req.Method, req.RequestURI)))
	} else {
		http.Error(res, "Function already exist\n"+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func ExtractFunction(res http.ResponseWriter, jsonBodyReq io.Reader) (name string, memory int, code, pack string) {
	var jsonBody interface{}
	err := json.NewDecoder(jsonBodyReq).Decode(&jsonBody)
	if err != nil {
		http.Error(res, err.Error(), 400)
		return
	}

	var bodyData = jsonBody.(map[string]interface{})
	return bodyData["name"].(string), int(bodyData["memory"].(float64)), bodyData["code"].(string), bodyData["package"].(string)
}

func functionDelete(res http.ResponseWriter, req *http.Request) {
	var name = strings.Split(req.RequestURI, "/")[2]

	if len(db.SelectFunction(name)) > 0 {
		db.DeleteFunction(name)
		res.Write([]byte(fmt.Sprintf("Function Deleted [%v] %v\n", req.Method, req.RequestURI)))
	} else {
		http.Error(res, "Function don't exist\n"+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}
