package gocoap

import (
	"fmt"
	"strings"
	"time"
)

// Example demonstrates how to use the gocoap package.
func Example() {
	// Create a new client with a 5-second timeout
	client := NewClient(5 * time.Second)

	// Example URL (this is a placeholder, replace with a real CoAP server URL)
	url := "coap://coap.me:5683/test"

	// Perform a GET request
	fmt.Println("Performing GET request to", url)
	response, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))

	// Perform a PUT request with a payload
	payload := strings.NewReader("Hello, CoAP!")
	url = "coap://coap.me:5683/large"
	fmt.Println("Performing PUT request to", url)
	response, err = client.Put(url, 0, payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))

	// Perform a POST request with a payload and JSON content format
	jsonPayload := strings.NewReader(`{"key":"value"}`)
	fmt.Println("Performing POST request with JSON content format to", url)
	response, err = client.Post(url, 50, jsonPayload) // 50 = application/json
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))

	// Perform a DELETE request
	fmt.Println("Performing DELETE request to", url)
	response, err = client.Delete(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))

	// Perform a GET request with non-confirmable message
	fmt.Println("Performing GET request with non-confirmable message to", url)
	response, err = client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))

	// Perform a PUT request with a payload, text content format, and confirmable message
	textPayload := strings.NewReader("Plain text payload")
	fmt.Println("Performing PUT request with text content format and confirmable message to", url)
	response, err = client.Put(url, 0, textPayload) // 0 = text/plain
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response))
}
