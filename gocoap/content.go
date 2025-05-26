// Package gocoap provides functionality for interacting with CoAP servers.
package gocoap

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
)

// ValidateContentType validates the given coap content coding id integer
// and returns a valid message.MediaType value.
// It returns an error if the media type is not valid.
// Common CoAP media types:
// 0: text/plain; charset=utf-8
// 40: application/link-format
// 41: application/xml
// 42: application/octet-stream
// 47: application/exi
// 50: application/json
// 60: application/cbor
func ValidateContentType(contentId int) (message.MediaType, error) {

	// Convert the integer to message.MediaType
	ct := message.MediaType(contentId)

	// Check if it's a registered media type (0-65535)
	if contentId >= 0 && contentId <= 65535 {
		// It's in the valid range
		return ct, nil
	}
	return 0, fmt.Errorf("invalid media type: %d", contentId)
}
