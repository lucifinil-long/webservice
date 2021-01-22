package webservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// webService provides web service
type webService struct {
	Config
	server           *http.Server
	templatesManager *templatesManager
	watcher          *fsnotify.Watcher
}

var (
	faviconHandler http.Handler
)

func init() {
	faviconHandler = http.StripPrefix("/", http.FileServer(http.Dir("./")))
}

func (ws *webService) PagesTempLatesDir() string {
	return ws.Config.PagesTempLatesDir
}

func (ws *webService) WidgetsTempLatesDir() string {
	return ws.Config.WidgetsTempLatesDir
}

func (ws *webService) TemplatesManager() TemplatesManager {
	return ws.templatesManager
}

// initAndServe initialize web service instance and start web service
// 	will panic if start web service failed
func (ws *webService) initAndServe(
	conf *Config) {
	// init member
	ws.Config = *conf
	if ws.Logger == nil {
		ws.Logger = &logger{}
	}

	ws.initTemplatesManager()

	webAddr := fmt.Sprintf("%v:%v", conf.WebAddr, conf.Port)
	// init http server
	ws.server = &http.Server{
		Addr:         webAddr,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		TLSConfig:    nil,
	}

	mux := http.NewServeMux()
	// add static directory
	for p, d := range conf.Statics {
		mux.Handle(p, http.StripPrefix(p, http.FileServer(http.Dir(d))))
	}
	// add handler func
	mux.HandleFunc("/", ws.dispatch)
	// add pprof invoke
	if conf.Pprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	ws.server.Handler = mux

	// start web service asynchronously
	go ws.run()
}

func (ws *webService) initTemplatesManager() {
	ws.templatesManager = buildTemplatesManager(ws.PagesTempLatesDir(), ws.PageGlobPattern,
		ws.WidgetsTempLatesDir(), ws.WidgetGlobPattern, ws.Logger)

	var err error
	ws.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		ws.Logger.Error("new watcher failed with", err)
		panic(err)
	}

	// start watcher events handler
	go ws.watcherEventsHandler()

	// watch directories
	pagesTemplatesDir := strings.TrimSpace(ws.PagesTempLatesDir())
	widgetsTempLatesDir := strings.TrimSpace(ws.WidgetsTempLatesDir())
	if pagesTemplatesDir != "" && ifDirExists(pagesTemplatesDir) {
		err = ws.watcher.Add(pagesTemplatesDir)
		if err != nil {
			ws.Logger.Error(pagesTemplatesDir, "is added watcher failed with", err)
			panic(err)
		}
	}
	if widgetsTempLatesDir != "" && ifDirExists(widgetsTempLatesDir) {
		err = ws.watcher.Add(widgetsTempLatesDir)
		if err != nil {
			ws.Logger.Error(widgetsTempLatesDir, "is added watcher failed with", err)
			panic(err)
		}
	}
}

func (ws *webService) watcherEventsHandler() {
	pagesTemplatesDir := strings.TrimSpace(ws.PagesTempLatesDir())
	pagePattern := strings.TrimSpace(ws.PageGlobPattern)
	for {
		select {
		case event, ok := <-ws.watcher.Events:
			if !ok {
				ws.Logger.Trace("fsnotify watcher is closed")
				return
			}
			ws.Logger.Trace("fsnotify watcher event", event)
			if event.Op&( /*fsnotify.Write|*/ fsnotify.Remove|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) > 0 &&
				pagePattern != "" && pagesTemplatesDir != "" {
				ws.Logger.Trace("refresh page templates of web service", ws.ServiceAddr())
				ws.templatesManager.Refresh(ws.PagesTempLatesDir(), ws.PageGlobPattern,
					ws.WidgetsTempLatesDir(), ws.WidgetGlobPattern)
			}

		case err, ok := <-ws.watcher.Errors:
			if !ok {
				ws.Logger.Trace("fsnotify watcher is closed")
				return
			}
			ws.Logger.Warn("fsnotify watcher error", err)
		}
	}
}

func (ws *webService) run() {
	var err error
	if strings.TrimSpace(ws.TLSCert) == "" ||
		strings.TrimSpace(ws.TLSKey) == "" {
		// there is not tls cert and key file info, just start http service
		ws.Logger.Trace("ListenAndServe for", ws.ServiceAddr())
		err = ws.server.ListenAndServe()
	} else {
		ws.Logger.Trace("ListenAndServeTLS for", ws.ServiceAddr())
		err = ws.server.ListenAndServeTLS(strings.TrimSpace(ws.TLSCert), strings.TrimSpace(ws.TLSKey))
	}

	if err != nil && err != http.ErrServerClosed {
		ws.Logger.Error("start web service", ws.ServiceAddr(), "with error", err)
		panic(err)
	}

	ws.Logger.Trace("web service", ws.ServiceAddr(), "is stopped", err)
}

func (ws *webService) Close() error {
	if ws.watcher != nil {
		ws.watcher.Close()
	}

	if ws.server == nil {
		return nil
	}

	// use context to control timeout of http.Server.Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := ws.server.Shutdown(ctx)
	if err != nil {
		ws.Logger.Error("http server shutdown with error", err)
	}
	return err
}

func (ws *webService) ServiceAddr() string {
	if ws.server == nil {
		return "service is not running"
	}
	return ws.server.Addr
}

func remoteAddrOfRequest(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	if xffh := r.Header.Get("X-Forwarded-For"); xffh != "" {
		remoteAddr = fmt.Sprintf("%v on behalf of %v", r.RemoteAddr, xffh)
	}

	return remoteAddr
}

func (ws *webService) dispatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	remoteAddr := remoteAddrOfRequest(r)

	ws.Logger.Trace("service", ws.server.Addr,
		"get request from", remoteAddr, "path", r.URL.Path, "form", r.Form, "post form", r.PostForm)

	if r.URL.Path == "/favicon.ico" {
		faviconHandler.ServeHTTP(w, r)
		ws.Logger.Trace("service", ws.server.Addr,
			"handled request from", remoteAddr, "path", r.URL.Path)
		return
	}

	rsp := ws.checkAuth(r.URL.Path, r.RemoteAddr)
	if rsp != nil {
		ws.Logger.Warn("service", ws.server.Addr, "checkAuth for",
			remoteAddr, "returned", rsp)
		ws.jsonResponseWithStatus(w, r, rsp, rsp.Status)
		return
	}

	// threat others as interface access
	handler := ws.handerForPath(r.URL.Path)
	if handler == nil {
		ws.jsonResponseWithStatus(w, r, &ServiceResponse{
			Status:  http.StatusBadRequest,
			Message: "bad request",
			Data:    map[int]int{},
		}, http.StatusBadRequest)
		ws.Logger.Warn("service", ws.server.Addr, "invalid path", r.URL.Path,
			"remote address", remoteAddr)
		return
	}

	resp := handler(w, r, ws)
	if resp != nil {
		if resp.StatusCode > 0 && resp.StatusCode != 200 {
			ws.jsonResponseWithStatus(w, r, resp, resp.StatusCode)
		} else {
			ws.jsonResponse(w, r, resp)
		}
	} else {
		ws.Logger.Trace("service", ws.server.Addr, "remote address", remoteAddr,
			"path", r.URL.Path, "handler has responsed by itself")
	}
}

func (ws *webService) handerForPath(path string) RequestHandlerFunc {
	if v, ok := ws.Handlers[path]; ok {
		return v
	}

	// allows case-insenstitve path
	for k, v := range ws.Handlers {
		if strings.EqualFold(path, k) {
			return v
		}
	}

	ws.Logger.Warn("service", ws.server.Addr, "not found handler for", path)
	return nil
}

func jsonResponse(w http.ResponseWriter, r *http.Request, data *ServiceResponse, host string, logger Logger) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	var dataBytes []byte
	var err error

	if data.Indent {
		dataBytes, err = json.MarshalIndent(data, "", "    ")
	} else {
		dataBytes, err = json.Marshal(data)
	}

	remoteAddr := remoteAddrOfRequest(r)

	if err != nil {
		jsonRsp := `{"errno":500,"errmsg":"internal error","Data":{}}`
		logger.Error("service", host,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with json marshal error", err, "response data", jsonRsp)
		fmt.Fprint(w, jsonRsp)
		return
	}

	rspData := string(dataBytes)
	_, err = fmt.Fprintf(w, "%v", rspData)

	var logData interface{}
	logData = data
	if logger.CheckLevel(logLevelDebug) {
		logData = rspData
	}

	if err != nil {
		logger.Error("service", host,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with error", err, "response data", logData)
	} else {
		logger.Trace("service", host,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with data", logData)
	}
}

func (ws webService) jsonResponse(w http.ResponseWriter, r *http.Request, data *ServiceResponse) {
	jsonResponse(w, r, data, ws.server.Addr, ws.Logger)
}

func (ws webService) jsonResponseWithStatus(w http.ResponseWriter, r *http.Request, data *ServiceResponse, status int) {
	remoteAddr := remoteAddrOfRequest(r)
	ws.Logger.Trace("service", ws.server.Addr,
		"response request from", remoteAddr, "path", r.RequestURI, "with status", status)
	w.WriteHeader(status)
	ws.jsonResponse(w, r, data)
}
