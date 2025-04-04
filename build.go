package go_connectwise

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"sync"
)

// NOTE - this one took some hand-tweaking because the ConnectWise API spec is invalid and inconsistent.
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./oapi-cw-config.yaml ./cw-api.json

var (
	once         sync.Once
	singleClient *ClientOptions
	singletonErr error
)

type ClientOptions struct {
	baseURL    string
	clientId   string
	company    string
	publicKey  string
	privateKey string
	client     *http.Client
}

var clientOnce sync.Once
var cwClient *ClientWithResponses

// ConnectToConnectWise3 - initializes the connection to the ConnectWise API, providing all arguments
func ConnectToConnectWise3(baseURL, company, clientId, publicKey, privateKey string) *ClientWithResponses {
	clientOnce.Do(func() {
		log.Debug("Connecting to connectwise API with provided arguments")
		options, err := CreateOrGetClient(baseURL, company, clientId, publicKey, privateKey)
		if err != nil {
			log.Fatal(err)
		}
		cwApi, err := NewClientWithResponses(options.baseURL,
			WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
				options.SetHeaders(req)
				return nil
			}))
		if err != nil {
			log.Fatal(err)
		}
		log.Debug(cwApi)
		log.Info("Connected to connectwise API")
		cwClient = cwApi
	})
	return cwClient
}

// ConnectToConnectWise - initializes the connection to the ConnectWise API, loading from the environment variables
func ConnectToConnectWise() *ClientWithResponses {
	clientOnce.Do(func() {
		cwClient = connectToConnectWise2()
	})
	return cwClient
}

func connectToConnectWise2() *ClientWithResponses {
	log.Debug("Connecting to connectwise API")
	options, _ := CreateOrGetClient("", "", "", "", "") // Load from environment variables if not provided
	cwApi, err := NewClientWithResponses("https://na.myconnectwise.net/v4_6_release/apis/3.0",
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			options.SetHeaders(req)
			return nil
		}))
	if err != nil {
		log.Fatal(err)
	}
	log.Debug(cwApi)
	log.Info("Connected to connectwise API")

	return cwApi
}

func CreateOrGetClient(baseUrl, company, clientId, publicKey, privateKey string) (*ClientOptions, error) {
	once.Do(func() {
		// Get required environment variables
		if baseUrl == "" {
			baseUrl = os.Getenv("CW_BASE_URL")
		}
		if company == "" {
			company = os.Getenv("CW_COMPANY")
		}
		if clientId == "" {
			clientId = os.Getenv("CW_CLIENT_ID")
		}
		if publicKey == "" {
			publicKey = os.Getenv("CW_PUBLIC_KEY")
		}
		if privateKey == "" {
			privateKey = os.Getenv("CW_PRIVATE_KEY")
		}

		// Validate required environment variables
		if baseUrl == "" || company == "" || publicKey == "" || privateKey == "" || clientId == "" {
			singletonErr = fmt.Errorf("missing required environment variables. Please ensure CW_BASE_URL, CW_COMPANY, CW_CLIENT_ID, CW_PUBLIC_KEY, and CW_PRIVATE_KEY are set")
			return
		}

		// Create the client
		singleClient = NewClientOptions(baseUrl, company, clientId, publicKey, privateKey)
	})

	if singletonErr != nil {
		return nil, singletonErr
	}

	return singleClient, nil
}

func NewClientOptions(baseURL, company, clientId, publicKey, privateKey string) *ClientOptions {
	return &ClientOptions{
		baseURL:    baseURL,
		company:    company,
		clientId:   clientId,
		publicKey:  publicKey,
		privateKey: privateKey,
		client:     &http.Client{},
	}
}

func (c *ClientOptions) SetHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.connectwise.com+json; version=2024.13")
	req.Header.Set("clientId", c.clientId)
	// Create the authentication string: company+publicKey:privateKey
	auth := fmt.Sprintf("%s+%s:%s", c.company, c.publicKey, c.privateKey)
	// Encode to base64
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
}

func GetViewUrl2[T any](x *T) (string, error) {
	if x == nil {
		return "", nil
	}
	return GetViewUrl((interface{})(*x))
}

func GetViewUrl(x interface{}) (string, error) {
	switch v := x.(type) {
	case Contact:
		contact := v
		return fmt.Sprintf("https://na.myconnectwise.net/v4_6_release/services/system_io/router/openrecord.rails?locale=en_US&recordType=ContactFV&companyName=nexigen&recid=%d", *contact.Id), nil
	case Company:
		company := v
		return fmt.Sprintf("https://na.myconnectwise.net/v4_6_release/services/system_io/router/openrecord.rails?locale=en_US&recordType=CompanyFV&recid=%d&companyName=nexigen", *company.Id), nil
	case Ticket:
		ticket := v
		return fmt.Sprintf("https://na.myconnectwise.net/v4_6_release/services/system_io/Service/fv_sr100_request.rails?service_recid=%d&companyName=nexigen", *ticket.Id), nil
	case Agreement:
		agreement := v
		return fmt.Sprintf("https://na.myconnectwise.net/v4_6_release/services/system_io/router/openrecord.rails?recordType=AgreementFV&recid=%d&companyName=nexigen", *agreement.Id), nil
	case Project:
		project := v
		return fmt.Sprintf("https://na.myconnectwise.net/v4_6_release/services/system_io/router/openrecord.rails?recordType=ProjectHeaderFV&recid=%d&companyName=nexigen", *project.Id), nil
	default:
		return "", fmt.Errorf("getViewUrl: Unknown type: %T", x)
	}
}
