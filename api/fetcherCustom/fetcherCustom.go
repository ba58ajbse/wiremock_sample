package fetchercustom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	PROTECTED_PATH = "/protected"
	LOGIN_PATH     = "/login"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type CredentialProvider interface {
	GetCredentials(ctx context.Context, paramName string) (map[string]string, error)
}

type SSMCredentialProvider struct {
	SSMClient *ssm.Client
}

func (p *SSMCredentialProvider) GetCredentials(ctx context.Context, paramName string) (map[string]string, error) {
	resp, err := p.SSMClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var credentials map[string]string
	if err := json.Unmarshal([]byte(*resp.Parameter.Value), &credentials); err != nil {
		return nil, err
	}
	return credentials, nil
}

type Fetcher struct {
	Client            Doer
	Domain            string
	Cookie            *http.Cookie
	ParamName         string
	CredentialFetcher CredentialProvider
}

func NewFetcher(domain, paramName string, client Doer, cp CredentialProvider) *Fetcher {
	if client == nil {
		client = http.DefaultClient
	}
	return &Fetcher{
		Client:            client,
		Domain:            domain,
		ParamName:         paramName,
		CredentialFetcher: cp,
	}
}

func (f *Fetcher) login(ctx context.Context) error {
	credentials, err := f.CredentialFetcher.GetCredentials(ctx, f.ParamName)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(credentials)

	url := f.Domain + LOGIN_PATH

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: status %d", resp.StatusCode)
	}

	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" { // 仮にCookie名をauth_tokenとする
			f.Cookie = c
			return nil
		}
	}

	return fmt.Errorf("auth cookie not found")
}

func (f *Fetcher) Fetch(ctx context.Context, payload any) (*Response, error) {
	if f.Cookie == nil {
		if err := f.login(ctx); err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := f.Domain + PROTECTED_PATH

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(f.Cookie)

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

// Example structures
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

func main() {
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	ssmClient := ssm.NewFromConfig(awsCfg)
	credentialProvider := &SSMCredentialProvider{SSMClient: ssmClient}

	fetcher := NewFetcher(
		"https://api.example.com",
		"/myapp/params",
		nil,
		credentialProvider,
	)

	payload := map[string]string{"data": "example"}
	resp, err := fetcher.Fetch(ctx, payload)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %+v\n", resp)
}
