package fetcher

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type UserInfo struct {
	Username string `json:"username"`
}

type Response struct {
	Meta Meta
	Data Data
}

type Meta struct {
	Error string `json:"error,omitempty"`
}
type Data struct {
	Message string `json:"message,omitempty"`
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Fetcher struct {
	Client   HttpClient
	Endpoint string
}

func NewFetcher(endpoint string, client HttpClient) *Fetcher {
	if client == nil {
		client = http.DefaultClient
	}
	return &Fetcher{
		Client:   client,
		Endpoint: endpoint,
	}
}

func (f *Fetcher) Fetch(payload any) (*Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, f.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
