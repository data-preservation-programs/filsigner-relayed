package client

import (
	"fmt"
	"github.com/data-preservation-programs/filsigner-relayed/model"
)

type RequestError struct {
	StatusCode model.StatusCode
	Message    string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("Request failed with status code %d (%s): %s", e.StatusCode, model.StatusCodeString[e.StatusCode], e.Message)
}
