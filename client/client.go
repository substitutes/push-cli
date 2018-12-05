package client

import (
	"io"
	"net/http"
)

type Client struct {
	Client  *http.Client
	Request *http.Request
}

// UserAgent contains the user agent used for the push CLI HTTP client
const UserAgent = "PushCLI/0.1 github.com/substitutes/push-cli"

// NewAuthClient creates an authenticated client
func New(url, method string, body io.Reader) (Client, error) {
	c := &http.Client{}
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return Client{}, err
	}
	// Set a custom UA
	r.Header.Set("User-Agent", UserAgent)

	return Client{Client: c, Request: r}, nil
}
