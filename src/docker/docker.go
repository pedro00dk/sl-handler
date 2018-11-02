package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/orisano/uds"
)

const (
	dockerSocketPath = "/var/run/docker.sock"
)

// Client provides methods for accessing docker host
type Client struct {
	unixHTTPClient *http.Client
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
func (c *Client) CreateImage(name, code, pack, dockerfile string) {
	var buffer bytes.Buffer
	tarWriter := tar.NewWriter(&buffer)
	files := []struct{ name, body string }{
		{"code.js", code},
		{"package.json", pack},
		{"Dockerfile", dockerfile},
	}
	for _, file := range files {
		tarHeader := &tar.Header{Name: file.name, Mode: 0600, Size: int64(len(file.body))}
		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			log.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(file.body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		log.Fatal(err)
	}
	tarWriter.Close()
	response, err := c.unixHTTPClient.Post(fmt.Sprintf("http://docker/build?t=%v", name), "application/x-tar", &buffer)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, response.Body)
	//fmt.Println(response.Body)
}
