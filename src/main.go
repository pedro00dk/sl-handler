package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"./database"
	"./docker"
)

var (
	db = database.Database{}
)

func testDockerClient() {
	client := docker.Client{}
	client.Init()
	if isConnected := client.IsConnected(); !isConnected {
		fmt.Println("Failed to connect")
	}

	imageName := "function:function"
	dockerfile, _ := ioutil.ReadFile("./dockerfiles/node/Dockerfile")
	serverJs, _ := ioutil.ReadFile("./dockerfiles/node/server.js")
	codeJs, _ := ioutil.ReadFile("./dockerfiles/node/code.js")
	createImageTime := client.CreateImage(
		imageName,
		docker.FileInfo{Name: "Dockerfile", Text: string(dockerfile)},
		docker.FileInfo{Name: "server.js", Text: string(serverJs)},
		docker.FileInfo{Name: "code.js", Text: string(codeJs)},
	)
	fmt.Println(createImageTime)
	containerID, createContainerTime := client.CreateContainer(imageName)
	fmt.Println(createContainerTime)
	time.Sleep(1 * time.Second)
	fmt.Println(client.StartContainer(containerID))
	time.Sleep(1 * time.Second)
	fmt.Println(client.StopContainer(containerID))
	time.Sleep(1 * time.Second)
	fmt.Println(client.DeleteContainer(containerID))
	time.Sleep(1 * time.Second)
	fmt.Println(client.DeleteImage(imageName))
}

func main() {
	db.Connect()
	//testDockerClient()

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
