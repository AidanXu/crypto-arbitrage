package binance

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
    BaseURL = "https://testnet.binance.vision"
)

type Client struct {
    HTTPClient *http.Client
    APIKey     string
    SecretKey  string
}

func NewClient() *Client {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    return &Client{
        HTTPClient: http.DefaultClient,
        APIKey:     os.Getenv("API_KEY"),
        SecretKey:  os.Getenv("SECRET_KEY"),
    }
}

func (c *Client) DoGetRequest(endpoint string, params map[string]string) ([]byte, error) {
    // Sort and encode parameters
    values := url.Values{}
    for key, value := range params {
        values.Add(key, value)
    }
    queryString := values.Encode()

    req, err := http.NewRequest("GET", BaseURL+endpoint+"?"+queryString, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Add("X-MBX-APIKEY", c.APIKey)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return body, nil
}

func (c *Client) DoPostRequest(endpoint string, payload string) ([]byte, error) {
    req, err := http.NewRequest("POST", BaseURL+endpoint, strings.NewReader(payload))
    if err != nil {
        return nil, err
    }

    req.Header.Add("X-MBX-APIKEY", c.APIKey)
    req.Header.Add("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return body, nil
}