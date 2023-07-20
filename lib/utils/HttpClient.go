package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type HttpClient struct {
	client  *http.Client
	headers map[string]string
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: time.Second * 30, // Timeout after 30 seconds
		},
		headers: make(map[string]string),
	}
}

func NewPodmanSocketClient() *HttpClient {
	return NewHttpClient().WithUnixSocket(os.Getenv("XDG_RUNTIME_DIR") + "/podman/podman.sock")
}

func (c *HttpClient) WithUnixSocket(unixSocketPath string) *HttpClient {
	var transport http.RoundTripper

	if unixSocketPath != "" {
		transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", unixSocketPath)
			},
		}

		c.client.Transport = transport
	}

	return c
}

func (c *HttpClient) WithHeaders(headers map[string]string) *HttpClient {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

func (c *HttpClient) Get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *HttpClient) Post(url string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *HttpClient) Pool(urls []string) ([]string, error) {
	var wg sync.WaitGroup
	responses := make([]string, len(urls))
	errors := make([]error, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			body, err := c.Get(url)
			if err != nil {
				errors[i] = err
				return
			}
			responses[i] = string(body)
		}(i, url)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return responses, nil
}
