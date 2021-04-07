package ipfsapi

import "net/http"

// Client is a client for the IPFS HTTP API.
type Client struct {
	c    *http.Client
	base string
}

// NewClient returns a new Client created with the given options.
func NewClient(options ...ClientOption) *Client {
	c := Client{
		c:    http.DefaultClient,
		base: "http://localhost:8080",
	}
	for _, option := range options {
		option(&c)
	}

	return &c
}

type ClientOption func(*Client)

// WithHTTPClient uses the given http.Client instead of
// http.DefaultClient.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.c = client
	}
}

// WithBaseURL sets the base URL for accessing the API. The default
// is "http://localhost:8080".
func WithBaseURL(base string) ClientOption {
	return func(c *Client) {
		c.base = base
	}
}
