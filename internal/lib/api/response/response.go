package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string      `json:"status"`
	Error  interface{} `json:"error,omitempty"`
	Alert  interface{} `json:"alert,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
	StatusAlert = "Alert"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg interface{}) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func Alert(msg interface{}) Response {
	return Response{
		Status: StatusAlert,
		Alert:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
