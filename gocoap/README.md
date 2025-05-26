# gocoap

A CoAP client tool similar to node package coap-cli

* Command line program to interact with CoAP servers
* Accept a sub-command, options, and a target URL
  * Example command invocation:
    * `gocoap [options] <sub-command> <url>`
* Sub commands:
  * `get` - performs a GET request
  * `put` - performs a PUT request
  * `post` - performs a POST request
  * `delete` - performs a DELETE request
* Options:
  * `-t <duration>` - request timeout (default: 5s)
  * `-p <payload>` - payload for PUT/POST requests
  * `-f <file>` - file containing payload for PUT/POST requests
  * `-n` - use non-confirmable messages (default: confirmable)
  * `-c <content-type>` - content format for requests (text, json, xml, octet, link)
  * `-b <option>` -- block size option
  * `-x` - print request time
  * `-v` - verbose output
  * `-q` - quiet: do not print status codes
  * `-h` - help: print help
* Features:
  * Support for both confirmable and non-confirmable messages
  * Support for different content formats
  * Robust URL parsing with support for query parameters
  * Verbose output option for debugging
* Uses the github.com/plgd-dev/go-coap/v3/coap package
* Uses Go standard libraries where possible
* Primary code is in the tools/gocoap directory
* CLI tool main function is in the tools/cmd/gocoap directory

For detailed usage and examples, see the [CLI tool README](../cmd/gocoap/README.md).
