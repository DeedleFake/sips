package ipfsapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is a client for the IPFS HTTP API.
type Client struct {
	client *http.Client
	base   string
}

// NewClient returns a new Client created with the given options.
func NewClient(options ...ClientOption) *Client {
	c := Client{
		client: http.DefaultClient,
		base:   "http://localhost:8080",
	}
	for _, option := range options {
		option(&c)
	}

	return &c
}

func (c *Client) post(data interface{}, endpoint string, args url.Values) error {
	url := c.base + "/" + endpoint + "?" + args.Encode()

	rsp, err := c.client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("post to %q: %w", endpoint, err)
	}
	defer rsp.Body.Close()

	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	err = json.Unmarshal(buf, data)
	if err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

type ID struct {
	Addresses       []string
	AgentVersion    string
	ID              string
	ProtocolVersion string
	Protocols       []string
	PublicKey       string
}

func (c *Client) ID() (ID, error) {
	var id ID
	err := c.post(&id, "/api/v0/id", nil)
	return id, err
}

type PinAdd struct {
	Pins     []string
	Progress int
}

func (c *Client) PinAdd(cids ...string) (PinAdd, error) {
	var data PinAdd
	err := c.post(&data, "/api/v0/pin/add", url.Values{
		"arg":      cids,
		"progress": []string{"true"},
	})
	return data, err
}

type PinLS struct {
	CID  string
	Type PinType
}

// PinLS gets information about pins from the node.
func (c *Client) PinLS(pintype PinType, cids ...string) ([]PinLS, error) {
	var data struct {
		Keys map[string]struct {
			Type PinType
		}
	}
	err := c.post(&data, "/api/v0/pin/ls", url.Values{
		"type": []string{string(pintype)},
		"arg":  cids,
	})
	if err != nil {
		return nil, err
	}

	pins := make([]PinLS, 0, len(data.Keys))
	for cid, info := range data.Keys {
		pins = append(pins, PinLS{
			CID:  cid,
			Type: info.Type,
		})
	}

	return pins, nil
}

func (c *Client) SwarmConnect(addr string) error {
	var data struct{}
	return c.post(&data, "/api/v0/swarm/connect", url.Values{
		"arg": []string{addr},
	})
}

type ClientOption func(*Client)

// WithHTTPClient uses the given http.Client instead of
// http.DefaultClient.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

// WithBaseURL sets the base URL for accessing the API. The default
// is "http://localhost:5001".
func WithBaseURL(base string) ClientOption {
	return func(c *Client) {
		c.base = base
	}
}

type PinType string

const (
	Direct    PinType = "direct"
	Indirect  PinType = "indirect"
	Recursive PinType = "recursive"
	All       PinType = "all"
)
