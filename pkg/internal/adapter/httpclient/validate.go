package httpclient

import (
	"fmt"
	"regexp"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

var headerValuePattern = regexp.MustCompile(`\r|\n`)

func (hp *HTTPClientAdapter[T]) validateHeaderValue(value string) error {
	if headerValuePattern.MatchString(value) {
		err := fmt.Errorf("header contains unsupported character")
		hp.notifyHTTPClientError(err)
		hp.NotifyLoggers(types.ErrorLevel, "httpclient invalid header value: %v", err)
		return err
	}
	return nil
}
