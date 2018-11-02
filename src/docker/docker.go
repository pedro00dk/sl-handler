package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
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
func (c *Client) CreateImage(name string, files ...FileInfo) (time.Duration, error) {
	startTime := time.Now()

	buffer := bytes.Buffer{}
	tarWriter := tar.NewWriter(&buffer)

	for _, file := range files {
		tarHeader := &tar.Header{Name: file.Name, Mode: 0600, Size: int64(len(file.Text))}
		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			return time.Since(startTime), err
		}
		if _, err := tarWriter.Write([]byte(file.Text)); err != nil {
			return time.Since(startTime), err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return time.Since(startTime), err
	}

	tarWriter.Close()
	response, err := c.unixHTTPClient.Post(fmt.Sprintf("http://docker/build?t=%v", name), "application/x-tar", &buffer)
	if err != nil {
		return time.Since(startTime), err
	}
	io.Copy(os.Stdout, response.Body)
	return time.Since(startTime), err
}
