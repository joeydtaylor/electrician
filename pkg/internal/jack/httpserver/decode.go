package httpserver

import (
	"net/http"

	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

func (h *httpServerAdapter[T]) parseRequest(r *http.Request) (T, error) {
	var decoded T
	if err := utils.DecodeJSON(r.Body, &decoded); err != nil {
		return decoded, err
	}
	return decoded, nil
}
