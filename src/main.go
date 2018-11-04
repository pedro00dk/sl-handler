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
	db                            = database.Database{}
	mdb                           = database.NewMetricBD("../metrics.json")
	mdbMetricChan, mdbPersistChan = mdb.StartMetricDBRoutine()
	dockerClient                  = docker.Client{}
	dockerfile, _                 = ioutil.ReadFile("../dockerfiles/node/Dockerfile")
	serverJS, _                   = ioutil.ReadFile("../dockerfiles/node/server.js")
)

func main() {
	db.Connect()
	dockerClient.Init()
	if isConnected := dockerClient.IsConnected(); !isConnected {
		fmt.Println("Failed to connect")
	}

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
		dockerClient.CreateImage(
			name,
			docker.FileInfo{Name: "Dockerfile", Text: string(dockerfile)},
			docker.FileInfo{Name: "server.js", Text: string(serverJS)},
			docker.FileInfo{Name: "package.json", Text: pack},
			docker.FileInfo{Name: "code.js", Text: code},
		)
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
		dockerClient.DeleteImage(name)
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
	requestData := req.RequestURI[6:]
	slashIndex := strings.Index(requestData, "/")
	if slashIndex == -1 {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte("Function endpoint not provided"))
		return
	}
	imageName := requestData[:slashIndex]

	containerID, containerCreateTime := dockerClient.CreateContainer(imageName)
	fmt.Printf("## Container ID: %v\n", containerID)
	fmt.Printf("## Create Container Time: %v\n", containerCreateTime)

	containerIP, containerStartTime := dockerClient.StartContainer(containerID)
	fmt.Printf("## Container IP: %v\n", containerIP)
	fmt.Printf("## Start Container Time: %v\n", containerStartTime)

	startApplicationConnectionTime := time.Now()
	var applicationRunTime time.Duration
	gatewayReq, err := http.NewRequest(req.Method, fmt.Sprintf("http://%v:8080/%v", containerIP, requestData[len(imageName)+1:]), req.Body)
	var gatewayRes *http.Response
	for i := 0; i < 200; i++ {
		fmt.Printf("Connection tries: %v\n", i)
		startApplicationRunTime := time.Now()
		gatewayRes, err = http.DefaultClient.Do(gatewayReq)
		fmt.Println(err)
		if err == nil {
			applicationRunTime = time.Since(startApplicationRunTime)
			fmt.Printf("## Request Run Time: %v\n", applicationRunTime)
			fmt.Println("Success!")
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	applicationConnectionTime := time.Since(startApplicationConnectionTime)
	fmt.Printf("## Request Time: %v\n", applicationConnectionTime)

	applicationCode := gatewayRes.StatusCode
	applicationBody, _ := ioutil.ReadAll(gatewayRes.Body)
	res.WriteHeader(applicationCode)
	res.Write(applicationBody)

	containerStopTime := dockerClient.StopContainer(containerID)
	containerDeleteTime := dockerClient.DeleteContainer(containerID)
	fmt.Printf("## Stop Container Time: %v\n", containerStopTime)
	fmt.Printf("## Delete Container Time: %v\n", containerDeleteTime)
	// fmt.Println(client.DeleteImage(imageName))

	metric := database.Metric{
		Function:                  imageName,
		ContainerID:               containerID,
		ContainerCreateTime:       containerCreateTime,
		ContainerStartTime:        containerStartTime,
		ApplicationConnectionTime: applicationConnectionTime,
		ApplicationRunTime:        applicationRunTime,
		ApplicationCode:           applicationCode,
		ContainerStopTime:         containerStopTime,
		ContainerDeleteTime:       containerDeleteTime,
	}

	mdbPersistChan <- true // disable later
	mdbMetricChan <- metric
}

// func serialize() {
// 	   mdbPersistChan <- true
// 	   mdbMetricChan <- database.Metric{}
// }
