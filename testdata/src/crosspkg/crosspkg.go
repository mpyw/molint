package crosspkg

import (
	"bytes"
	"strings"

	"lib"
)

func Constructor() *lib.Client { // want Constructor:"nonnil r0"
	return lib.NewClient()
}

func Maybe() *lib.Client {
	return lib.Lookup("x") // want `from a call to lib.Lookup, which may return nil`
}

func Dialed() (*lib.Client, error) { // want Dialed:"nonnil r0"
	c, err := lib.Dial("x")
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Global() *lib.Client { // want Global:"nonnil r0"
	return lib.Default
}

func Unset() *lib.Client {
	return lib.Unset // want `from package variable Unset`
}

func Generic() *int { // want Generic:"nonnil r0"
	return lib.Generic[int]()
}

func Std() *strings.Reader { // want Std:"nonnil r0"
	return strings.NewReader("x")
}

func StdBuffer() *bytes.Buffer { // want StdBuffer:"nonnil r0"
	return bytes.NewBufferString("x")
}
