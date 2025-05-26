# gocoap CLI Tool

A command-line tool for interacting with CoAP servers, similar to the node package coap-cli.

## Installation

```bash
go install github.com/larryr/tools/cmd/gocoap@latest
```

## Usage

```
gocoap <command> [options] <url>
```

### Commands

- `get` - Perform a GET request
- `put` - Perform a PUT request
- `post` - Perform a POST request
- `delete` - Perform a DELETE request

### Options

- `-t <duration>` - Request timeout (default: 5s)
- `-p <payload>` - Payload for PUT/POST requests
- `-f <file>` - File containing payload for PUT/POST requests
- `-n` - Use non-confirmable messages (default: confirmable)
- `-c <format>` - Content format for requests (text, json, xml, octet, link)
- `-v` - Verbose output

## Examples

### GET Request

```bash
gocoap get coap://example.org:5683/test
```

### PUT Request with Payload

```bash
gocoap put -p "Hello, CoAP!" coap://example.org:5683/test
```

### POST Request with Payload from File

```bash
gocoap post -f payload.txt coap://example.org:5683/test
```

### DELETE Request

```bash
gocoap delete coap://example.org:5683/test
```

### Custom Timeout

```bash
gocoap get -t 10s coap://example.org:5683/test
```

### Non-Confirmable Message

```bash
gocoap get -n coap://example.org:5683/test
```

### Specify Content Format

```bash
gocoap put -p '{"key":"value"}' -c json coap://example.org:5683/test
```

### Verbose Output

```bash
gocoap get -v coap://example.org:5683/test
```

### Combining Options

```bash
gocoap post -f data.json -c json -n -v -t 15s coap://example.org:5683/resource
```

## Development

The gocoap CLI tool is built on top of the [github.com/plgd-dev/go-coap/v3](https://github.com/plgd-dev/go-coap) package and provides a simple interface for interacting with CoAP servers.

To build the tool from source:

```bash
git clone https://github.com/larryr/tools.git
cd tools
go build -o gocoap ./cmd/gocoap
```
