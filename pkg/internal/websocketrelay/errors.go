package websocketrelay

import "fmt"

var errMissingToken = fmt.Errorf("missing bearer token")
var errMissingHeaders = fmt.Errorf("missing headers")

func errMissingHeader(k string) error {
	return fmt.Errorf("missing/invalid header %s", k)
}
