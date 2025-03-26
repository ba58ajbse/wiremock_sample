package fetchercustom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockCredentialProvider struct {
	mock.Mock
}

func (m *MockCredentialProvider) GetCredentials(ctx context.Context, paramName string) (map[string]string, error) {
	args := m.Called(ctx, paramName)
	return args.Get(0).(map[string]string), args.Error(1)
}

func TestFetcher_Login(t *testing.T) {
	ctx := context.Background()

	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "token123"})
		w.WriteHeader(http.StatusOK)
	}))
	defer loginServer.Close()

	mockCredProvider := &MockCredentialProvider{}
	mockCredProvider.On("GetCredentials", mock.Anything, "param").Return(map[string]string{"user_name": "test", "password": "pass"}, nil)

	fetcher := NewFetcher(loginServer.URL, "param", nil, mockCredProvider)

	err := fetcher.login(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher.Cookie)
	assert.Equal(t, "auth_token", fetcher.Cookie.Name)
	assert.Equal(t, "token123", fetcher.Cookie.Value)
}
