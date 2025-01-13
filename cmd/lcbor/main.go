package main

import (
	"flag"
	"fmt"
	"github.com/larryr/tools/lcbor/cborstream"
	"io"
	"os"

	"github.com/larryr/tools/lcbor/test"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: cat data.cbor | lcbor -d\n")
}

var encode = flag.Bool("e", false, "encode data")
var decode = flag.Bool("d", false, "decode data")
var tester = flag.Bool("test", false, "generate test cbor encoded test data")

func main() {
	flag.Usage = usage

	flag.Parse()
	if flag.NArg() != 0 {
		usage()
	}

	if err := lcbor(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func lcbor(in io.Reader, out io.Writer) error {

	if *decode {
		return cborstream.DecodeToOutput(in, out)
	}
	if *tester {
		return test.GenerateTestData(out)
	}
	return nil
}
