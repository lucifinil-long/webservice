package webservice

import (
	"path/filepath"
	"time"
)

// Config stores web service config
type Config struct {
	WebAddr                  string
	Port                     uint16
	Statics                  map[string]string
	Handlers                 map[string]RequestHandlerFunc
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	AuthMap                  map[string]map[string]int
	PagesTempLatesDir        string
	PageGlobPattern          string
	WidgetsTempLatesDir      string
	WidgetGlobPattern        string
	TemplatesRefreshInterval time.Duration
	Logger                   Logger
	Pprof                    bool
	TLSCert                  string
	TLSKey                   string
	UploadsDir               string
}

// BuildConfig builds a default http config which can be convert to https config easy
func BuildConfig() *Config {
	return &Config{
		WebAddr:                  "0.0.0.0",
		Port:                     80,
		ReadTimeout:              10 * time.Minute,
		WriteTimeout:             10 * time.Minute,
		PagesTempLatesDir:        filepath.Join(getCurrentDirectory(), "templates/pages"),
		PageGlobPattern:          "*.*",
		WidgetsTempLatesDir:      filepath.Join(getCurrentDirectory(), "templates/widgets"),
		WidgetGlobPattern:        "*.*",
		TemplatesRefreshInterval: time.Second * 60,
		Logger:                   &logger{},
		Pprof:                    true,
		UploadsDir:               filepath.Join(getCurrentDirectory(), "uploads"),
	}
}

// SetLogger set logger to config
func (conf *Config) SetLogger(log interface{}) {
	if conf.Logger == nil {
		conf.Logger = &logger{}
	}

	if log == nil {
		return
	}

	l := log.(Logger)
	if l != nil {
		conf.Logger = l
	}
}
