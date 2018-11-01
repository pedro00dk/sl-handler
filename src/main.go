package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	// server := &http.Server{
	// 	Addr:           ":8000",
	// 	Handler:        Handler{},
	// 	ReadTimeout:    10 * time.Second,
	// 	WriteTimeout:   10 * time.Second,
	// 	MaxHeaderBytes: 1 << 20,
	// }
	// server.ListenAndServe()

	http.HandleFunc("/function", function)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/call", call)
	http.ListenAndServe(":8000", nil)
}

// Handler represents the http struct that hold a function to process requests.
// type Handler struct{}

// func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
// 	res.Write([]byte(fmt.Sprintf("[%v] %v", req.Method, req.RequestURI)))
// }

func function(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
		return
	}

	if req.Body == nil {
		http.Error(res, "Empty body", 400)
		return
	}

	var body interface{}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(res, err.Error(), 400)
		return
	}

	var bodyData = body.(map[string]interface{})
	fmt.Println(bodyData["action"])
	fmt.Println(bodyData["name"])
	fmt.Println(bodyData["code"])
	fmt.Println(bodyData["package"])
	var containerOptions = bodyData["container-options"].(map[string]interface{})
	fmt.Println(containerOptions["cpus"])
	fmt.Println(containerOptions["memory"])

	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}
