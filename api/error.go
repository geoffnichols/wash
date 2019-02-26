package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// This approach was adapted from https://blog.golang.org/error-handling-and-go

// ErrorFields represents the fields of an ErrorObj
type ErrorFields = map[string]interface{}

// ErrorObj represents an API error object
type ErrorObj struct {
	Kind   string      `json:"kind"`
	Msg    string      `json:"msg"`
	Fields ErrorFields `json:"fields"`
}

func (e *ErrorObj) Error() string {
	jsonBytes, err := json.Marshal(e)
	if err != nil {
		// We should never hit this code-path, but better safe than sorry
		return fmt.Sprintf("Kind: %v, Msg: %v, Fields: %v", e.Kind, e.Msg, e.Fields)
	}

	return string(jsonBytes)
}

func newErrorObj(kind string, message string, fields ErrorFields) *ErrorObj {
	return &ErrorObj{
		Kind:   "puppetlabs.wash/" + kind,
		Msg:    message,
		Fields: fields,
	}
}

func newUnknownErrorObj(err error) *ErrorObj {
	return newErrorObj("unknown-error", err.Error(), ErrorFields{})
}

func newStreamingErrorObj(reason string) *ErrorObj {
	return newErrorObj("streaming-error", reason, ErrorFields{})
}

// ErrorResponse represents an error response
type errorResponse struct {
	statusCode int
	body       *ErrorObj
}

func (e *errorResponse) Error() string {
	return e.body.Error()
}

// Below are all of Wash's API error responses

func unknownErrorResponse(err error) *errorResponse {
	statusCode := http.StatusInternalServerError
	body := newUnknownErrorObj(err)

	return &errorResponse{statusCode, body}
}

func entryNotFoundResponse(path string, reason string) *errorResponse {
	fields := ErrorFields{"path": path}

	statusCode := http.StatusNotFound
	body := newErrorObj(
		"entry-not-found",
		fmt.Sprintf("Could not find entry %v: %v", path, reason),
		fields,
	)

	return &errorResponse{statusCode, body}
}

func pluginDoesNotExistResponse(plugin string) *errorResponse {
	fields := ErrorFields{"plugin": plugin}

	statusCode := http.StatusNotFound
	body := newErrorObj(
		"plugin-does-not-exist",
		fmt.Sprintf("Plugin %v does not exist", plugin),
		fields,
	)

	return &errorResponse{statusCode, body}
}

func unsupportedActionResponse(path string, action *action) *errorResponse {
	fields := ErrorFields{
		"path":   path,
		"action": action,
	}

	statusCode := http.StatusNotFound
	msg := fmt.Sprintf("Entry %v does not support the %v action: It does not implement the %v protocol", path, action.Name, action.Protocol)
	body := newErrorObj(
		"unsupported-action",
		msg,
		fields,
	)

	return &errorResponse{statusCode, body}
}

func badRequestResponse(path string, reason string) *errorResponse {
	fields := ErrorFields{"path": path}
	body := newErrorObj(
		"bad-request",
		fmt.Sprintf("Bad request on %v: %v", path, reason),
		fields,
	)
	return &errorResponse{http.StatusBadRequest, body}
}

func erroredActionResponse(path string, action *action, reason string) *errorResponse {
	fields := ErrorFields{
		"path":   path,
		"action": action.Name,
	}

	statusCode := http.StatusInternalServerError
	body := newErrorObj(
		"errored-action",
		fmt.Sprintf("The %v action errored on %v: %v", action.Name, path, reason),
		fields,
	)

	return &errorResponse{statusCode, body}
}

func httpMethodNotSupported(method string, path string, supported []string) *errorResponse {
	fields := ErrorFields{
		"method":    method,
		"path":      path,
		"supported": supported,
	}

	body := newErrorObj(
		"http-method-not-supported",
		fmt.Sprintf("The %v method is not supported for %v, supported methods are: %v", method, path, strings.Join(supported, ", ")),
		fields,
	)

	return &errorResponse{http.StatusNotFound, body}
}