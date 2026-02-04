package helper

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
)

type APIClient struct {
	client  *resty.Client
	baseURL string
}

type RestyResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Error      error
}

// NewAPIClient initializes a new Resty client with default configurations
func NewAPIClient(baseURL string) *APIClient {
	client := resty.New().
		SetTimeout(30*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(2*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	return &APIClient{
		client:  client,
		baseURL: baseURL,
	}
}

func (r *APIClient) SetBasicAuth(username, password string) *APIClient {
	r.client.SetBasicAuth(username, password)
	return r
}

// SetHeader allows adding custom headers to the client
func (r *APIClient) SetHeader(key, value string) *APIClient {
	r.client.SetHeader(key, value)
	return r
}

// SetTimeout allows customizing the timeout
func (r *APIClient) SetTimeout(timeout time.Duration) *APIClient {
	r.client.SetTimeout(timeout)
	return r
}

// Get performs a GET request
func (r *APIClient) Get(endpoint string) *RestyResponse {
	resp, err := r.client.R().
		Get(r.baseURL + endpoint)

	return r.buildResponse(resp, err)
}

// Post performs a POST request with a body
func (r *APIClient) Post(endpoint string, body interface{}) *RestyResponse {
	resp, err := r.client.R().
		SetBody(body).
		Post(r.baseURL + endpoint)

	return r.buildResponse(resp, err)
}

// Put performs a PUT request with a body
func (r *APIClient) Put(endpoint string, body interface{}) *RestyResponse {
	resp, err := r.client.R().
		SetBody(body).
		Put(r.baseURL + endpoint)

	return r.buildResponse(resp, err)
}

// Delete performs a DELETE request
func (r *APIClient) Delete(endpoint string) *RestyResponse {
	resp, err := r.client.R().
		Delete(r.baseURL + endpoint)

	return r.buildResponse(resp, err)
}

// buildResponse constructs a standardized response
func (r *APIClient) buildResponse(resp *resty.Response, err error) *RestyResponse {
	response := &RestyResponse{
		Headers: make(map[string]string),
	}

	if err != nil {
		response.Error = err
		return response
	}

	if resp == nil {
		response.Error = errors.New("nil response received")
		return response
	}

	response.StatusCode = resp.StatusCode()
	response.Body = resp.Body()

	// Copy headers
	for key, values := range resp.Header() {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	return response
}

// GetJSON performs a GET request and unmarshals the response into the provided struct
func (r *APIClient) GetJSON(endpoint string, result interface{}) error {
	resp := r.Get(endpoint)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.StatusCode >= 400 {
		return errors.New("request failed with status: " + string(resp.StatusCode))
	}
	return json.Unmarshal(resp.Body, result)
}

// PostJSON performs a POST request and unmarshals the response into the provided struct
func (r *APIClient) PostJSON(endpoint string, body, result interface{}) error {
	resp := r.Post(endpoint, body)
	if resp.Error != nil {
		return resp.Error
	}
	if resp.StatusCode >= 400 {
		return errors.New("request failed with status: " + string(resp.StatusCode))
	}
	return json.Unmarshal(resp.Body, result)
}
