package cborstream

import (
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"
	"github.com/fxamacker/cbor/v2"
)

func DecodeToOutput(in io.Reader, out io.Writer) error {
	dec := cbor.NewDecoder(in)
	for {
		var v interface{}
		err := dec.Decode(&v)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%#v\n\n", v)
		spew.Fdump(out, v)
	}
	return nil
}
