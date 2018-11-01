package main

import (
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
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}
