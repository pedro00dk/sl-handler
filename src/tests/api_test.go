package test

import (
	"testing"
	"io/ioutil"
	"github.com/ricardorobson/sl-handler/src/database"
	"github.com/ricardorobson/sl-handler/src/docker"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
	"io"
	"fmt"
)

var (
	db                            = database.Database{}
	mdb                           = database.NewMetricBD("../metrics.json")
	mdbMetricChan, mdbPersistChan = mdb.StartMetricDBRoutine()
	dockerClient                  = docker.Client{}
	dockerfile, _                 = ioutil.ReadFile("../dockerfiles/node/Dockerfile")
	serverJS, _                   = ioutil.ReadFile("../dockerfiles/node/server.js")
)

type API struct {
	Client  *http.Client
	baseURL string
}

func TestInsert(t *testing.T) {
	prepareDatabase(t)
	server := prepareServer(t)
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	api := API{server.Client(), server.URL}

	body, err := api.DoInsert()
	
	if  err != nil {
		t.Errorf("API Error")
	}

	if !bytes.Equal(body, []byte(string(http.StatusCreated))){
		t.Errorf("Response error")
	}

	cleanDatabase(t)

}

func TestDelete(t *testing.T) {
	prepareDatabase(t)
	server := prepareServer(t)
	// Close the server when test finishes
	defer server.Close()
	populateDB(t)

	// Use Client & URL from our local test server
	api := API{server.Client(), server.URL}
	body, err := api.DoDelete()
	
	if  err != nil {
		t.Errorf("API Error")
	}

	if !bytes.Equal(body, []byte("Function Deleted")){
		t.Errorf("Response error")
	}

	cleanDatabase(t)

}

func (api *API) DoInsert() ([]byte, error) {
	resp, err := api.Client.Post(
		api.baseURL + "/function",
		"application/json",
	 	bytes.NewBuffer([]byte(`{
			"name": "functest",
			"code": "module.exports.helloWorld = (req, res) => {\n    res.send(\"hello world\")\n}",
			"package": "{}",
			"memory": 200
		}`)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (api *API) DoDelete() ([]byte, error) {
	resp, err := api.Client.Post(
		api.baseURL + "/delete",
		"application/json",
	 	bytes.NewBuffer([]byte(`{
			"name": "functest"
		}`)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// http server that simules the API
func prepareServer(t *testing.T) (*httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		
		if req.URL.String() == "/function"{
			name, memory, code, pack := ExtractFunction(rw, req.Body)
			if len(db.SelectFunction(name)) == 0 {				
				dockerClient.CreateImage(
					name,
					docker.FileInfo{Name: "Dockerfile", Text: string(dockerfile)},
					docker.FileInfo{Name: "server.js", Text: string(serverJS)},
					docker.FileInfo{Name: "package.json", Text: pack},
					docker.FileInfo{Name: "code.js", Text: code},
				)
				
				db.InsertFunction(name, memory, code, pack)
				rw.Write([]byte(string(http.StatusCreated)))
			} else {
				http.Error(rw, "Function already exist\n"+http.StatusText(http.StatusConflict), http.StatusConflict)
			}
		} else if req.URL.String() == "/delete" {
			name := ExtractName(rw, req.Body)
			dockerClient.DeleteImage(name)
			var success = db.DeleteFunction(name)
			if success {
				rw.Write([]byte("Function Deleted"))
			} else{
				rw.Write([]byte(fmt.Sprintf("Cannot Delete function %v\n", name)))
			}

		} else {
			t.Errorf("Insert an URL well formatted")
		}
		
	}))

	return server
}

func prepareDatabase(t *testing.T) {
	db.Connect()
	dockerClient.Init()
	if isConnected := dockerClient.IsConnected(); !isConnected {
		t.Errorf("Failed to Connect")
	}
	
}

func populateDB(t *testing.T) {
	db.InsertFunction(
		"functest",
		 200,
	 	"module.exports.helloWorld = (req, res) => {\n    res.send(\"hello world\")\n}",
		  "{}")		  
}

func cleanDatabase(t *testing.T) {
	// TO-DO: MAKE IT DYNAMIC
	db.DeleteFunction("functest")
	db.Close()
}

// util
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

func ExtractName(res http.ResponseWriter, jsonBodyReq io.Reader) (name string) {
	var jsonBody interface{}
	err := json.NewDecoder(jsonBodyReq).Decode(&jsonBody)
	if err != nil {
		http.Error(res, err.Error(), 400)
		return
	}

	var bodyData = jsonBody.(map[string]interface{})
	return bodyData["name"].(string)
}

func functionGetByName(argument string) string {
	return string(db.SelectFunction(argument))
}