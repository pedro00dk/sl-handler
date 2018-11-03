package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"./docker"
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
	testDockerClient()

	http.HandleFunc("/function", function)
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/call", call)
	http.ListenAndServe(":8000", nil)
}

func function(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func metrics(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}

func call(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("[%v] %v\n", req.Method, req.RequestURI)))
}
