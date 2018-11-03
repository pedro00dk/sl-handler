package main

import (
	"bytes"
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

func main() {
	db.Connect()

	http.HandleFunc("/function/", function)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/call/", call)
	http.ListenAndServe(":8000", nil)
}

func function(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		functionGet(res, req)
	case "POST":
		functionPost(res, req)
	case "DELETE":
		functionDelete(res, req)
	default:
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func functionPost(res http.ResponseWriter, req *http.Request) {
	name, memory, code, pack := ExtractFunction(res, req.Body)
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

func functionGet(res http.ResponseWriter, req *http.Request) {
	var functions []database.Function
	if !strings.EqualFold(strings.Split(req.RequestURI, "/")[2], "") {
		var name = strings.Split(req.RequestURI, "/")[2]
		functions = db.SelectFunction(name)
	} else {
		functions = db.SelectAllFunction()
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(functions)
	res.Write(buf.Bytes())
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	client := docker.Client{}
	client.Init()
	if isConnected := client.IsConnected(); !isConnected {
		fmt.Println("Failed to connect")
	}
	imageName := "function:function"
	dockerfile, _ := ioutil.ReadFile("../dockerfiles/node/Dockerfile")
	serverJs, _ := ioutil.ReadFile("../dockerfiles/node/server.js")
	codeJs, _ := ioutil.ReadFile("../dockerfiles/node/code.js")
	createImageTime := client.CreateImage(
		imageName,
		docker.FileInfo{Name: "Dockerfile", Text: string(dockerfile)},
		docker.FileInfo{Name: "server.js", Text: string(serverJs)},
		docker.FileInfo{Name: "code.js", Text: string(codeJs)},
	)
	fmt.Printf("## Create Image Time: %v\n", createImageTime)

	containerID, createContainerTime := client.CreateContainer(imageName)
	fmt.Printf("## Container ID: %v\n", containerID)
	fmt.Printf("## Create Container Time: %v\n", createContainerTime)

	containerIP, startContainerTime := client.StartContainer(containerID)
	fmt.Printf("## Container IP: %v\n", containerIP)
	fmt.Printf("## Start Container Time: %v\n", startContainerTime)

	requestTime := time.Now()
	gatewayReq, err := http.NewRequest(req.Method, fmt.Sprintf("http://%v:8080/%v/", containerIP, req.RequestURI[6:]), req.Body)
	var gatewayRes *http.Response
	for i := 0; i < 200; i++ {
		fmt.Printf("Connection tries: %v\n", i)
		requestRunTime := time.Now()
		gatewayRes, err = http.DefaultClient.Do(gatewayReq)
		if err == nil {
			fmt.Printf("## Request Run Time: %v\n", time.Since(requestRunTime))
			fmt.Println("Success!")
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Printf("## Request Time: %v\n", time.Since(requestTime))

	code := gatewayRes.StatusCode
	body, _ := ioutil.ReadAll(gatewayRes.Body)
	res.WriteHeader(code)
	res.Write(body)

	stopContainerTime := client.StopContainer(containerID)
	deleteContainerTime := client.DeleteContainer(containerID)
	fmt.Printf("## Stop Container Time: %v\n", stopContainerTime)
	fmt.Printf("## Delete Container Time: %v\n", deleteContainerTime)
	// fmt.Println(client.DeleteImage(imageName))
}
