package webservice

import (
	"net/http"
	"time"
)

// RequestHandlerFunc defines union http request handler in this frame
//  if handler returned nil, frame will not response client;
//	otherwise frame will response client with json from return value
type RequestHandlerFunc func(http.ResponseWriter, *http.Request, []byte) *ServiceResponse

// WebService povides a web service interface definition
type WebService interface {
	// Close close web service
	Close() error
	// ServiceAddr returns service address
	ServiceAddr() string
}

// Logger defines logger interface
type Logger interface {
	// SetLevel
	SetLevel(uint)
	CheckLevel(uint) bool
	Write(string, bool, ...interface{})
	Debug(...interface{})
	Trace(...interface{})
	Warn(...interface{})
	Error(...interface{})
}

// BuildWebServiceAndServe build a WebService object and serve immediately
func BuildWebServiceAndServe(webAddr string,
	handlers map[string]RequestHandlerFunc,
	statics map[string]string,
	readTimeout, writeTimeout time.Duration,
	authMap map[string]map[string]int,
	logger Logger,
	tlsCertAndKey ...string) WebService {
	service := &webService{}

	service.initAndServe(webAddr, handlers, statics, readTimeout, writeTimeout, authMap, logger, tlsCertAndKey...)
	return service
}
