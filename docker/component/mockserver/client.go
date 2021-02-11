package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// QueryParameter is a query parameter for mockserver.
type QueryParameter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Response is an untyped response from mockserver expectation.
type Response struct {
	Status int         `json:"statusCode"`
	Body   interface{} `json:"body"`
}

// CallTimes configures the call times for an expectation.
type CallTimes struct {
	RemainingTimes int  `json:"remainingTimes"`
	Unlimited      bool `json:"unlimited"`
}

// Request is the request that an expectation matches against.
type Request struct {
	Method          string           `json:"method"`
	Path            string           `json:"path"`
	QueryParameters []QueryParameter `json:"queryStringParameters,omitempty"`
	Body            interface{}      `json:"body,omitempty"`
}

// WithJSONBody returns a Request with a JSON body set.
func (r Request) WithJSONBody(body interface{}) Request {
	type jsonBody struct {
		Type      string      `json:"type"`
		MatchType string      `json:"matchType"`
		JSON      interface{} `json:"json"`
	}

	r.Body = jsonBody{
		Type:      "json",
		MatchType: "STRICT",
		JSON:      body,
	}

	return r
}

// Expectation represents an expectation in mockserver.
type Expectation struct {
	Request  Request   `json:"httpRequest"`
	Response Response  `json:"httpResponse"`
	Times    CallTimes `json:"times"`
}

// Client is a client for mockserver.
type Client struct {
	host   string
	client *http.Client
}

// NewClient creates a new client.
func NewClient(address string) *Client {
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	return &Client{
		host:   address,
		client: http.DefaultClient,
	}
}

// CreateExpectation creates a new expectation (request/response) in the mockserver.
func (c *Client) CreateExpectation(expectation Expectation) error {
	reqBody, err := json.Marshal(expectation)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.host+"/expectation", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("201 status expected but %d received", resp.StatusCode)
	}

	return nil
}

// Reset deletes all expectations and recorded requests in mockserver.
func (c *Client) Reset() error {
	req, err := http.NewRequest(http.MethodPut, c.host+"/reset", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("200 status expected but %d received", resp.StatusCode)
	}

	return nil
}
