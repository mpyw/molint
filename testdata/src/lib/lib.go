package lib

type Client struct{}

func NewClient() *Client { // want NewClient:"nonnil r0"
	return &Client{}
}

func Lookup(name string) *Client {
	if name == "" {
		return nil // want `from a nil constant`
	}
	return &Client{}
}

func Dial(addr string) (*Client, error) { // want Dial:"nonnil r0"
	return &Client{}, nil
}

var Default = NewClient() // want Default:"nonnil"

var Unset *Client

func Generic[T any]() *T { // want Generic:"nonnil r0"
	return new(T)
}
