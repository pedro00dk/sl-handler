package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/orisano/uds"
)

const (
	dockerSocketPath = "/var/run/docker.sock"
)

// Client provides methods for accessing docker host
type Client struct {
	unixHTTPClient *http.Client
}

// FileInfo specifies the file name and its content
type FileInfo struct {
	Name string
	Text string
}

// Init starts a http socket client for the unix domain socket interface
func (c *Client) Init() {
	c.unixHTTPClient = uds.NewClient(dockerSocketPath)
}

// IsConnected checks if the connection was established
func (c *Client) IsConnected() bool {
	return c.unixHTTPClient != nil
}

// CreateImage creates a docker image with the received files, returns the time to create
func (c *Client) CreateImage(name string, files ...FileInfo) time.Duration {
	startTime := time.Now()

	tarBuffer := bytes.Buffer{}
	tarWriter := tar.NewWriter(&tarBuffer)
	for _, file := range files {
		tarHeader := &tar.Header{Name: file.Name, Mode: 0600, Size: int64(len(file.Text))}
		tarWriter.WriteHeader(tarHeader)
		tarWriter.Write([]byte(file.Text))
	}
	tarWriter.Close()

	response, _ := c.unixHTTPClient.Post(
		fmt.Sprintf("http://docker/build?t=%v", name),
		"application/x-tar",
		&tarBuffer,
	)
	response.Body.Close()

	return time.Since(startTime)
}

// CreateContainer initializes a container with the received image, returns the time to create and the container id
func (c *Client) CreateContainer(image string) (string, time.Duration) {
	startTime := time.Now()

	response, _ := c.unixHTTPClient.Post(
		"http://docker/containers/create",
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{ "Image": "%v" }`, image))),
	)
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// simple json result (json parsing is expensive -> use simple string manipulation)
	// var json map[string]interface{}
	// json.Unmarshal(createResponseBody, &createResponseJSON)
	// containerID := createResponseJSON["Id"].(string)

	containerID := string(body[7:71])
	fmt.Println(body)
	fmt.Println(string(containerID))

	return containerID, time.Since(startTime)
}

// StartContainer starts the container with the received containerID, returns the time to start the container
func (c *Client) StartContainer(containerID string) time.Duration {
	startTime := time.Now()

	response, _ := c.unixHTTPClient.Post(
		fmt.Sprintf("http://docker/containers/%v/start", containerID),
		"application/json",
		nil,
	)
	response.Body.Close()

	return time.Since(startTime)
}

// StopContainer stops the container with the received container Id, returns the time to stop
func (c *Client) StopContainer(containerID string) time.Duration {
	startTime := time.Now()

	response, _ := c.unixHTTPClient.Post(
		fmt.Sprintf("http://docker/containers/%v/kill", containerID),
		"application/json",
		nil,
	)
	response.Body.Close()

	return time.Since(startTime)
}

// DeleteContainer deletes the container with the received container Id, returns the time to stop
func (c *Client) DeleteContainer(containerID string) time.Duration {
	startTime := time.Now()

	request, _ := http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://docker/containers/%v", containerID),
		nil,
	)
	response, _ := c.unixHTTPClient.Do(request)
	response.Body.Close()

	return time.Since(startTime)
}
