package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Fetcher struct {
	Domain   string
	Endpoint string
	Client   *http.Client
}

func NewFetcher(url, endpoint string) Fetcher {
	return Fetcher{
		Domain:   url,
		Endpoint: endpoint,
		Client:   http.DefaultClient,
	}
}

func (f *Fetcher) fetch(payload, output any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	url := f.Domain + f.Endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		fmt.Println("リクエストエラー:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("status error code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(output)
}

func main() {
	fe := NewFetcher("http://localhost:8888", "/api/login")
	payload := UserInfo{
		Username: "admin",
	}

	output := Response{}

	err := fe.fetch(payload, &output)
	if err != nil {
		fmt.Println(err)
		return
	}

	// msg, code, err := check(body)
	// if err != nil {
	// 	fmt.Printf("エラーレスポンス -> msg: %v, code: %v\n", err.Error(), code)
	// 	return
	// }

	fmt.Printf("レスポンス -> msg: %v\n", output.Data.Message)
	// fmt.Printf("レスポンス -> msg: %v, code: %v\n", msg, code)
}

func check(data []byte) (msg string, code int, e error) {
	v := Response{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err.Error(), 500, err
	}

	if v.Meta.Error != "" {
		return v.Meta.Error, 500, fmt.Errorf("error: %s", v.Meta.Error)
	}

	return v.Data.Message, 200, nil
}
