package webservice

import (
	"net/http"
	"net/url"
)

// RequestHandlerFunc defines union http request handler in this frame
//  if handler returned nil, frame will not response client;
//	otherwise frame will response client with json from return value
type RequestHandlerFunc func(http.ResponseWriter, *http.Request, WebService) *ServiceResponse

// WebService defines a web service interface definition
type WebService interface {
	// Close close web service
	Close() error
	// ServiceAddr returns service address
	ServiceAddr() string
	// PagesTempLatesDir returns page templates directory of web service
	PagesTempLatesDir() string
	// WidgetsTempLatesDir returns widget templates directory of web service
	WidgetsTempLatesDir() string
	// TemplatesManager get service related templates manager
	TemplatesManager() TemplatesManager
}

// TemplatesManager defines templates manager interface definition
type TemplatesManager interface {
	// RenderTemplate render template to http.ResponseWriter
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
}

// StartWebService starts a web service with config
func StartWebService(conf *Config) WebService {
	service := &webService{}
	service.initAndServe(conf)
	return service
}

// HTTPProxy defines http proxy interface
type HTTPProxy interface {
	ForwardRequest(w http.ResponseWriter, r *http.Request, target *url.URL)
	AgentRequest(r *http.Request, target *url.URL) (*http.Response, error)
}

// BuildHTTPProxy builds http proxy object
func BuildHTTPProxy(logger interface{}) HTTPProxy {
	return &proxy{logger: ConvertLoggerMust(logger)}
}

// // BuildTemplatesManager builds a templates manager object
// func BuildTemplatesManager(
// 	pagesTemplateDir, pagePattern,
// 	widgetsTemplateDir, widgetPattern string, log Logger) (TemplatesManager, error) {
// 	return buildTemplatesManager(pagesTemplateDir, pagePattern,
// 		widgetsTemplateDir, widgetPattern, log)
// }
