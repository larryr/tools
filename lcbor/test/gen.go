package test

import (
	"io"

	"github.com/fxamacker/cbor/v2"
)

// GenerateTestData will output a cbor encoded test data to the provided io.Writer
func GenerateTestData(out io.Writer) error {

	enc := cbor.NewEncoder(out)

	obj := LcborTestDataObject{
		AString: "object string",
		AInt:    32000,
	}
	data := LcborTestData{
		IsBool:     true,
		IsString:   "string",
		IsInt32Sm:  -1,
		IsFloat32:  3.14,
		IsInt32Lg:  1000000,
		IsFloat64:  3.14159265359,
		IsInt64:    1000000000000,
		IsIntArray: []int{1, 2, 3, 4, 5},
		IsObject:   obj,
	}

	err := enc.Encode(data)
	return err
}

type LcborTestDataObject struct {
	AString string `cbor:"aString"`
	AInt    int    `cbor:"aInt"`
}
type LcborTestData struct {
	IsBool     bool                `cbor:"b1"`
	IsString   string              `cbor:"nam"`
	IsInt32Sm  int32               `cbor:"v1"`
	IsFloat32  float32             `cbor:"f1"`
	IsInt32Lg  int32               `cbor:"v2"`
	IsFloat64  float64             `cbor:"f2"`
	IsInt64    int64               `cbor:"v3"`
	IsIntArray []int               `cbor:"a1"`
	IsObject   LcborTestDataObject `cbor:"o1"`
}
