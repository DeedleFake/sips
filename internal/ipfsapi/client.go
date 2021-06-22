package ipfsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func (c *Client) post(ctx context.Context, data interface{}, endpoint string, args url.Values) error {
	rsp, err := c.postResponse(ctx, endpoint, args)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer rsp.Body.Close()

	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %v: %q", rsp.Status, buf)
	}

	err = json.Unmarshal(buf, data)
	if err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) postResponse(ctx context.Context, endpoint string, args url.Values) (*http.Response, error) {
	url := c.base + "/api/v0/" + endpoint + "?" + args.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post to %q: %w", endpoint, err)
	}

	return rsp, err
}

type ID struct {
	Addresses       []string
	AgentVersion    string
	ID              string
	ProtocolVersion string
	Protocols       []string
	PublicKey       string
}

func (c *Client) ID(ctx context.Context) (ID, error) {
	var id ID
	err := c.post(ctx, &id, "id", nil)
	return id, err
}

type PinAdd struct {
	Pins []string
}

func (c *Client) PinAdd(ctx context.Context, cids ...string) (PinAdd, error) {
	var data PinAdd
	err := c.post(ctx, &data, "pin/add", url.Values{
		"arg":      cids,
		"progress": []string{"false"},
	})
	return data, err
}

type PinAddProgress struct {
	Pins     []string
	Progress int
	Err      error
}

func (c *Client) PinAddProgress(ctx context.Context, cids ...string) (<-chan PinAddProgress, error) {
	rsp, err := c.postResponse(ctx, "pin/add", url.Values{
		"arg":      cids,
		"progress": []string{"true"},
	})
	if err != nil {
		return nil, err
	}

	progress := make(chan PinAddProgress)
	go func() {
		defer rsp.Body.Close()
		defer close(progress)

		d := json.NewDecoder(rsp.Body)
		for {
			if !d.More() {
				break
			}

			var data PinAddProgress
			err := d.Decode(&data)
			if err != nil {
				select {
				case <-ctx.Done():
				case progress <- PinAddProgress{Err: err}:
				}
				return
			}

			select {
			case <-ctx.Done():
				return
			case progress <- data:
			default:
			}
		}
	}()

	return progress, nil
}

type PinLs struct {
	CID  string
	Type PinType
}

// PinLs gets information about pins from the node.
func (c *Client) PinLs(ctx context.Context, pintype PinType, cids ...string) ([]PinLs, error) {
	var data struct {
		Keys map[string]struct {
			Type PinType
		}
	}
	err := c.post(ctx, &data, "pin/ls", url.Values{
		"type": []string{string(pintype)},
		"arg":  cids,
	})
	if err != nil {
		return nil, err
	}

	pins := make([]PinLs, 0, len(data.Keys))
	for cid, info := range data.Keys {
		pins = append(pins, PinLs{
			CID:  cid,
			Type: info.Type,
		})
	}

	return pins, nil
}

func (c *Client) PinUpdate(ctx context.Context, oldCID, newCID string, unpin bool) ([]string, error) {
	var data struct {
		Pins []string
	}
	err := c.post(ctx, &data, "pin/update", url.Values{
		"arg":   []string{oldCID, newCID},
		"unpin": []string{strconv.FormatBool(unpin)},
	})
	return data.Pins, err
}

func (c *Client) PinRm(ctx context.Context, cids ...string) ([]string, error) {
	var data struct {
		Pins []string
	}
	err := c.post(ctx, &data, "pin/rm", url.Values{
		"arg": cids,
	})
	return data.Pins, err
}

func (c *Client) SwarmConnect(ctx context.Context, addr string) error {
	var data struct{}
	return c.post(ctx, &data, "swarm/connect", url.Values{
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
		c.base = strings.TrimSuffix(base, "/")
	}
}

type PinType string

const (
	Direct    PinType = "direct"
	Indirect  PinType = "indirect"
	Recursive PinType = "recursive"
	All       PinType = "all"
)
