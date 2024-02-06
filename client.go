package dify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	host         string
	apiSecretKey string
	httpClient   *http.Client
}

func NewClientWithConfig(c *ClientConfig) *Client {
	var httpClient = &http.Client{}

	if c.Timeout != 0 {
		httpClient.Timeout = c.Timeout
	}
	if c.Transport != nil {
		httpClient.Transport = c.Transport
	}

	return &Client{
		host:         c.Host,
		apiSecretKey: c.ApiSecretKey,
		httpClient:   httpClient,
	}
}

func NewClient(host, apiSecretKey string) *Client {
	return NewClientWithConfig(&ClientConfig{
		Host:         host,
		ApiSecretKey: apiSecretKey,
	})
}

func (c *Client) SendRequest(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) SendJSONRequest(req *http.Request, res interface{}) error {
	resp, err := c.SendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		}
		err = json.NewDecoder(resp.Body).Decode(&errBody)
		if err != nil {
			return err
		}
		return fmt.Errorf("HTTP response error: [%v]%v", errBody.Code, errBody.Message)
	}

	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetHost() string {
	var host = strings.TrimSuffix(c.host, "/")
	return host
}

func (c *Client) GetApiSecretKey() string {
	return c.apiSecretKey
}

func (c *Client) Api() *Api {
	return &Api{
		c: c,
	}
}
