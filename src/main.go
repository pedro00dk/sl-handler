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
	db            = database.Database{}
	dockerClient  = docker.Client{}
	dockerfile, _ = ioutil.ReadFile("../dockerfiles/node/Dockerfile")
)

const (
	functionEndpoint = "/function/"
	metricsEndpoint  = "/metrics"
	callEndpoint     = "/call/"
	port             = ":8000"
)

func main() {
	db.Connect()
	dockerClient.Init()
	if isConnected := dockerClient.IsConnected(); !isConnected {
		fmt.Println("Failed to connect")
	}

	http.HandleFunc(functionEndpoint, function)
	http.HandleFunc(metricsEndpoint, metrics)
	http.HandleFunc(callEndpoint, call)
	http.ListenAndServe(port, nil)
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

func functionGet(res http.ResponseWriter, req *http.Request) {
	var argument = req.RequestURI[len(functionEndpoint):]
	if !strings.EqualFold(argument, "") {
		var function = functionGetByName(argument)
		res.Write([]byte(function))
	} else {
		var functions = functionGetAll()
		res.Write([]byte(functions))
	}
}

func functionGetAll() string {
	return string(db.SelectAllFunction())
}

func functionGetByName(argument string) string {
	return string(db.SelectFunction(argument))
}

func functionPost(res http.ResponseWriter, req *http.Request) {
	name, memory, code, pack := ExtractFunction(res, req.Body)
	if len(db.SelectFunction(name)) == 0 {
		dockerClient.CreateImage(
			name,
			docker.FileInfo{Name: "package.json", Text: pack},
			docker.FileInfo{Name: "code.js", Text: code},
			docker.FileInfo{Name: "Dockerfile", Text: string(dockerfile)},
		)
		db.InsertFunction(name, memory, code, pack)
		var function = functionGetByName(name)
		res.Write([]byte(function))
		res.Write([]byte(fmt.Sprintf("Function Created at %v%v\n", req.RequestURI, name)))
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
		dockerClient.DeleteImage(name)
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
	var imageName = strings.Split(req.RequestURI, "/")[2]
	containerID, createContainerTime := dockerClient.CreateContainer(imageName)
	fmt.Printf("## Container ID: %v\n", containerID)
	fmt.Printf("## Create Container Time: %v\n", createContainerTime)

	containerIP, startContainerTime := dockerClient.StartContainer(containerID)
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

	stopContainerTime := dockerClient.StopContainer(containerID)
	deleteContainerTime := dockerClient.DeleteContainer(containerID)
	fmt.Printf("## Stop Container Time: %v\n", stopContainerTime)
	fmt.Printf("## Delete Container Time: %v\n", deleteContainerTime)
	// fmt.Println(client.DeleteImage(imageName))
}
