package main

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/substitutes/push-cli/client"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func BuildPath(path string) string {
	return (*server).String() + "/api/" + path
}

func GetURL(path string) *url.URL {
	u, err := url.Parse(BuildPath(path))
	if err != nil {
		log.Fatal("Failed to create URL: ", err)
	}
	return u
}

type API struct{}

func (*API) get(path string) (*http.Response, error) {
	c, err := client.New((*server).String()+"/api/"+path, "GET", nil)

	if err != nil {
		log.Fatal("Failed to create client: ", err)
	}

	c.Request.SetBasicAuth(*username, *password)

	resp, err := c.Client.Do(c.Request)
	if err != nil {
		log.Warn("Failed to make request: ", err)
		return nil, err
	}
	return resp, err
}

func (api *API) pingAPI() {
	// Verify connection against server
	log.Debug("Attempting to ping server")
	response, err := api.get("ping")
	if err != nil {
		log.Fatal("Failed to make request to server: ", err)
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Failed to decode response: ", err)
	}
	var pingResponse struct{ Status string `json:"status"` }
	if err := json.Unmarshal(responseBytes, &pingResponse); err != nil {
		log.Fatal("Failed to read response: ", err, " got HTTP status ", response.Status)
	}
	if pingResponse.Status != "OK" {
		log.Fatal("Failed to get ping response from server. Expected ok, got ", pingResponse.Status)
	}
}

// Push a specific file to the API
func (api *API) push(file *os.File) (*http.Response, error) {

	// Read contents of file
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("Failed to read file: ", fileBytes)
		return &http.Response{}, err
	}

	// Create a buffer
	body := new(bytes.Buffer)

	// Create a new multipart writer with that buffer
	mpw := multipart.NewWriter(body)

	// Create the MPW
	part, err := mpw.CreateFormFile("push", filepath.Base(file.Name()))
	if err != nil {
		return &http.Response{}, err
	}

	// Write filebytes to mpw
	part.Write(fileBytes)
	if err := mpw.Close(); err != nil {
		return &http.Response{}, err
	}

	cl, err := client.New((*server).String()+"/api/push", "POST", body)

	if err != nil {
		log.Warn("Failed to create client")
		return &http.Response{}, err
	}

	cl.Request.Header.Set("Content-Type", mpw.FormDataContentType())
	cl.Request.ContentLength = int64(body.Len())
	cl.Client.Transport = &http.Transport{Proxy: http.ProxyURL(*proxyURL)}

	buf := body.Bytes()
	cl.Request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(buf)
		return ioutil.NopCloser(r), nil
	}

	cl.Request.SetBasicAuth(*username, *password)
	return cl.Client.Do(cl.Request)
}

func (api *API) pushFiles() {
	// Push all files
	filepath.Walk(*directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			log.Debugf("%v is a directory - skipping!", path)
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			log.Warnf("Failed to open file %s: %s", path, err)
			return err
		}
		defer file.Close()
		response, err := api.push(file)
		if err != nil {
			log.Warn("Failed to push file to server: ", err)
			return err
		}
		defer response.Body.Close()
		responseBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Warn("Failed to read body of response: ", err)
			return err
		}
		if response.StatusCode != 201 {
			var v struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}

			if err := json.Unmarshal(responseBytes, &v); err != nil {
				log.Warn("Failed to unmarshal response: ", err)
				return err
			}
			log.Warnf("Failed to upload file: %v (%+v)", v.Message, v.Error)
			return errors.New(v.Error)
		}
		var v []struct {
			UploadedAt int64  `json:"uploaded_at"`
			Name       string `json:"name"`
			Size       int64  `json:"size"`
		}

		if err := json.Unmarshal(responseBytes, &v); err != nil {
			log.Warn("Failed to unmarshal response: ", err)
			return err
		}
		log.Debugf("Uploaded files: %+v", v)
		return nil
	})
}
