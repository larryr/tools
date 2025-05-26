package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/larryr/tools/gocoap"
)

func usage() {
	qfprintf(os.Stderr, "Usage: gocoap [options] <command> <url>\n\n")
	qfprintf(os.Stderr, "Commands:\n")
	qfprintf(os.Stderr, "  get     Perform a GET request\n")
	qfprintf(os.Stderr, "  put     Perform a PUT request\n")
	qfprintf(os.Stderr, "  post    Perform a POST request\n")
	qfprintf(os.Stderr, "  delete  Perform a DELETE request\n")
	qfprintf(os.Stderr, "  observe Perform an observe request\n")
	qfprintf(os.Stderr, "  block   block\n")
	qfprintf(os.Stderr, "  example Execute the example\n")
	qfprintf(os.Stderr, "  version Print version\n\n")
	qfprintf(os.Stderr, "Options:\n")
	qfprintf(os.Stderr, "  -t <duration>      Request timeout (default: 5s)\n")
	qfprintf(os.Stderr, "  -p <payload>       Payload for PUT/POST requests\n")
	qfprintf(os.Stderr, "  -f <file>          File containing payload for PUT/POST requests\n")
	qfprintf(os.Stderr, "  -n                 Use non-confirmable messages\n")
	qfprintf(os.Stderr, "  -c <content-type>  Content format (coap content-type id)\n")
	qfprintf(os.Stderr, "  -b <option>        block size option\n")
	qfprintf(os.Stderr, "  -x                 print request time\n")
	qfprintf(os.Stderr, "  -v                 Verbose output\n")
	qfprintf(os.Stderr, "  -q                 quiet: do not print status codes of received messages\n")
	qfprintf(os.Stderr, "  -h                 Show this help message\n\n")
	qfprintf(os.Stderr, "Examples:\n")
	qfprintf(os.Stderr, "  gocoap get coap://example.org:5683/test\n")
	qfprintf(os.Stderr, "  gocoap put -p \"Hello, CoAP!\" coap://example.org:5683/test\n")
	qfprintf(os.Stderr, "  gocoap post -f payload.txt -c json coap://example.org:5683/test\n")
	qfprintf(os.Stderr, "  gocoap get -n -v coap://example.org:5683/test\n")
}

func main() {
	// Define flags
	timeout := flag.Duration("t", 5*time.Second, "request timeout")
	payload := flag.String("p", "", "payload for PUT/POST requests")
	payloadFile := flag.String("f", "", "file containing payload for PUT/POST requests")
	nonConfirmable := flag.Bool("n", false, "use non-confirmable messages")
	contentId := flag.Int("c", 0, "media type for requests (numeric code)")
	verbose := flag.Bool("v", false, "verbose output")

	// Custom usage function
	flag.Usage = usage

	// Parse flags
	flag.Parse()

	// Check if a command was provided
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	// Get the command and URL
	command := strings.ToLower(flag.Arg(0))

	if command == "example" {
		gocoap.Example()
		os.Exit(0)
	}

	if flag.NArg() < 2 {
		qfprintf(os.Stderr, "Error: URL is required\n")
		usage()
		os.Exit(1)
	}
	url := flag.Arg(1)

	// Create a new client
	client := gocoap.NewClient(*timeout)

	// validate media type
	ct, err := gocoap.ValidateContentType(*contentId)
	if err != nil {
		qfprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Determine the payload for PUT/POST requests
	var payloadReader io.ReadSeeker
	if *payloadFile != "" {
		file, err := os.Open(*payloadFile)
		if err != nil {
			qfprintf(os.Stderr, "Error opening payload file: %v\n", err)
			os.Exit(3)
		}
		defer func() { _ = file.Close() }()

		payloadReader = file
	} else if *payload != "" {
		payloadReader = strings.NewReader(*payload)
	}

	// Print verbose information if requested
	if *verbose {
		qfprintf(os.Stderr, "CoAP Client Configuration:\n")
		qfprintf(os.Stderr, "  URL: %s\n", url)
		qfprintf(os.Stderr, "  Method: %s\n", command)
		qfprintf(os.Stderr, "  Timeout: %s\n", *timeout)
		qfprintf(os.Stderr, "  Confirmable: %v\n", !*nonConfirmable)
		qfprintf(os.Stderr, "  Content Format: %v\n", *contentId)
		if *payload != "" {
			qfprintf(os.Stderr, "  Payload: %s\n", *payload)
		}
		if *payloadFile != "" {
			qfprintf(os.Stderr, "  Payload File: %s\n", *payloadFile)
		}
		qfprintf(os.Stderr, "\n")
	}

	// Execute the command
	var response []byte
	switch command {
	case "get":
		response, err = client.Get(url)
	case "put":
		response, err = client.Put(url, ct, payloadReader)
	case "post":
		response, err = client.Post(url, ct, payloadReader)
	case "delete":
		response, err = client.Delete(url)
	default:
		qfprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		usage()
		os.Exit(1)
	}

	// Handle errors
	if err != nil {
		qfprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Println(string(response))
}

func qfprintf(wr io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(wr, format, a...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
