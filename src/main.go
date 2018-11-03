package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	// server := &http.Server{
	// 	Addr:           ":8000",
	// 	Handler:        Handler{},
	// 	ReadTimeout:    10 * time.Second,
	// 	WriteTimeout:   10 * time.Second,
	// 	MaxHeaderBytes: 1 << 20,
	// }
	// server.ListenAndServe()
	/*
		fmt.Print(time.Now())
		for index := 0; index < 100000; index++ {
			database.InsertFunction("Nome",2, 1024, "Código","Package")
			if index%1000==0{
				fmt.Println(index)
			}
		}
		fmt.Print(time.Now())
	*/

	fmt.Print(db.SelectAllFunction())
	//db.Close()
	/*
		client := docker.Client{}
		client.Init()
		if isConnected := client.IsConnected(); !isConnected {
			fmt.Println("Failed to connect")
		}
		elapsedTime, err := client.CreateImage("some:function", "", "", "FROM ubuntu")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(elapsedTime)
	*/
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

func methodPost(res http.ResponseWriter, req *http.Request) {
	switch req.RequestURI {
	case "/function":
		functionPost(res, req)
		//case "/metric"
	}
}

func functionPost(res http.ResponseWriter, req *http.Request) {
	name, code, pack, containerOptions := ExtractFunction(res, req.Body)
	if db.SelectFunction(name) != nil {
		db.InsertFunction(name, 512, code, pack)
	}
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))

	http.Error(res, "Function exist", 500)
}
func ExtractFunction(res http.ResponseWriter, jsonBodyReq io.Reader) (name, code, pack, containerOptions string) {
	var jsonBody interface{}
	err := json.NewDecoder(jsonBodyReq).Decode(&jsonBody)
	if err != nil {
		http.Error(res, err.Error(), 400)
		return
	}

	var bodyData = jsonBody.(map[string]interface{})
	return bodyData["name"].(string), bodyData["code"].(string), bodyData["package"].(string), bodyData["container-options"].(string)
}

func function(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		//methodGet(res, req)
	case "POST":
		methodPost(res, req)
	case "DELETE":
		//methodDelete(res, req)
	}

	/*
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

		resp, err := dockerAPIClient.Get("http://unix/images/json")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(resp.Body)
		io.Copy(os.Stdout, resp.Body)
		resp.Body.Close()

		res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
	*/
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}
