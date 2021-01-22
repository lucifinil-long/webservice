package webservice

import (
	"encoding/json"
	"errors"
)

const (
	// ErrorCodeSuccess indicates success
	ErrorCodeSuccess = 1
)

// ServiceResponse defines union web service response
type ServiceResponse struct {
	Status     int         `json:"status"`
	Message    interface{} `json:"message"`
	Data       interface{} `json:"data"`
	Indent     bool        `json:"-"`
	StatusCode int         `json:"-"`
}

func (sr ServiceResponse) String() string {
	data, _ := json.Marshal(sr)
	return string(data)
}

// ServiceResponseData defines union web service response data for parsing
type ServiceResponseData struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (srd ServiceResponseData) String() string {
	data, _ := json.Marshal(srd)
	return string(data)
}

var (
	// ErrorNotFound defines not found error
	ErrorNotFound = errors.New("not found")
	// ErrorWrongMethod defines wrong method method error
	ErrorWrongMethod = errors.New("wrong method")
	// ErrorParsedYet defines parsed error
	ErrorParsedYet = errors.New("request has parsed by others")
	// ErrorInvalidArgument defines invalid argument
	ErrorInvalidArgument = errors.New("invalid argument")
)
