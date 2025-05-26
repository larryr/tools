// Package gocoap provides functionality for interacting with CoAP servers.
package gocoap

import (
	"context"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"io"
	"net/url"
	"strings"
	"time"

	coapudp "github.com/plgd-dev/go-coap/v3/udp"
	coapclient "github.com/plgd-dev/go-coap/v3/udp/client"
)

// Client represents a CoAP client.
type Client struct {
	timeout time.Duration
}

// NewClient creates a new CoAP client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		timeout: timeout,
	}
}

// Get performs a GET request to the specified URL.
// TODO add options
func (c *Client) Get(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Parse the URL to get the host and path
	host, path, err := parseURL(url)
	if err != nil {
		return nil, err
	}

	// Create a CoAP client connection
	conn, err := coapudp.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer connClose(conn)

	resp, err := conn.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Read the response body

	body, err := io.ReadAll(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return body, nil
}

// Post performs a POST request to the specified URL with the given payload.
func (c *Client) Post(url string, contentId message.MediaType, payload io.ReadSeeker) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Parse the URL to get the host and path
	host, path, err := parseURL(url)
	if err != nil {
		return nil, err
	}

	// Create a CoAP client connection
	conn, err := coapudp.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer connClose(conn)

	// Read payload if provided
	var payloadBytes []byte
	if payload != nil {
		payloadBytes, err = io.ReadAll(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %v", err)
		}
	}

	resp, err := conn.Post(ctx, path, contentId, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return body, nil
}

// Put performs a PUT request to the specified URL with the given payload.
func (c *Client) Put(url string, contentId message.MediaType, payload io.ReadSeeker) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Parse the URL to get the host and path
	host, path, err := parseURL(url)
	if err != nil {
		return nil, err
	}

	// Create a CoAP client connection
	conn, err := coapudp.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer connClose(conn)

	resp, err := conn.Put(ctx, path, contentId, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return body, nil
}

// Delete performs a DELETE request to the specified URL.
func (c *Client) Delete(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Parse the URL to get the host and path
	host, path, err := parseURL(url)
	if err != nil {
		return nil, err
	}

	// Create a CoAP client connection
	conn, err := coapudp.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer connClose(conn)

	resp, err := conn.Delete(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return body, nil
}

// parseURL parses a CoAP URL and returns the host and path.
func parseURL(rawURL string) (string, string, error) {
	// Check if the URL starts with coap:// or coaps://
	if !strings.HasPrefix(rawURL, "coap://") && !strings.HasPrefix(rawURL, "coaps://") {
		return "", "", fmt.Errorf("invalid CoAP URL (must start with coap:// or coaps://): %s", rawURL)
	}

	// Replace the scheme with http:// or https:// to use the net/url package
	var httpURL string
	if strings.HasPrefix(rawURL, "coap://") {
		httpURL = "http://" + rawURL[7:]
	} else {
		httpURL = "https://" + rawURL[8:]
	}

	// Parse the URL using the net/url package
	parsedURL, err := url.Parse(httpURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Extract the host and path
	host := parsedURL.Host
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	// Add query parameters to the path if present
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	return host, path, nil
}

func connClose(conn *coapclient.Conn) {
	_ = conn.Close()
}
