package fetcher

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	// DoFunc func(req *http.Request) (*http.Response, error)
	mock.Mock
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	// return m.DoFunc(req)
	args := m.Called(req)

	resp := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}

func TestFetcher_Fetch(t *testing.T) {
	// モックレスポンス
	resBody := `{"data":{"message":"hello"}}`
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(resBody)),
		Header:     make(http.Header),
	}

	m := new(mockClient)
	m.On("Do", mock.AnythingOfType("*http.Request")).Return(mockResp, nil)

	fetcher := &Fetcher{
		Client:   m,
		Endpoint: "http://dummy/api",
	}

	payload := UserInfo{Username: "test"}
	res, err := fetcher.Fetch(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Data.Message != "hello" {
		t.Errorf("unexpected message: got %s, want %s", res.Data.Message, "hello")
	}
}
