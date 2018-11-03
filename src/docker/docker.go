package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

// CreateImage creates a docker image with the received files
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
	io.Copy(os.Stdout, response.Body)

	return time.Since(startTime)
}

// StartContainer initializes a container with the received image
func (c *Client) StartContainer(image string, memory int) time.Duration {
	startTime := time.Now()

	createResponse, _ := c.unixHTTPClient.Post(
		"http://docker/containers/create",
		"application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{ "Image": "%v" }`, image))),
	)
	createResponseBody, _ := ioutil.ReadAll(createResponse.Body)
	fmt.Println(string(createResponseBody))

	var createResponseJSON map[string]interface{}
	json.Unmarshal(createResponseBody, &createResponseJSON)
	fmt.Println(createResponseJSON["Id"])
	containerID := createResponseJSON["Id"].(string)

	startResponse, err := c.unixHTTPClient.Post(
		fmt.Sprintf("http://docker/containers/%v/start", containerID),
		"application/json",
		bytes.NewReader([]byte{}),
	)
	if err != nil {
	}
	io.Copy(os.Stdout, startResponse.Body)

	return time.Since(startTime)
}
