package webservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/lucifinil-long/logging"
)

// webService provides web service
type webService struct {
	handlers map[string]RequestHandlerFunc
	authMap  map[string]map[string]int
	server   *http.Server
	logger   logging.Logger
}

var (
	faviconHandler http.Handler
)

func init() {
	faviconHandler = http.StripPrefix("/", http.FileServer(http.Dir("./")))
}

// initAndServe initialize web service instance and start web service
// 	will panic if start web service failed
func (ws *webService) initAndServe(
	webAddr string,
	handlers map[string]RequestHandlerFunc,
	statics map[string]string,
	readTimeout, writeTimeout time.Duration,
	authMap map[string]map[string]int,
	logger logging.Logger,
	tlsCertAndKey ...string) {
	// init member
	ws.handlers = handlers
	ws.authMap = authMap
	if logger != nil {
		ws.logger = logger
	} else {
		logger, err := logging.GetLogger("webservice", "main", nil)
		if err != nil {
			panic(err)
		}
		ws.logger = logger
	}

	// init http server
	ws.server = &http.Server{
		Addr:         webAddr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		TLSConfig:    nil,
	}

	mux := http.NewServeMux()
	// add static directory
	for p, d := range statics {
		mux.Handle(p, http.StripPrefix(p, http.FileServer(http.Dir(d))))
	}
	// add handler func
	mux.HandleFunc("/", ws.dispatch)

	ws.server.Handler = mux

	go func() {
		// start web service asynchronously
		var err error
		if len(tlsCertAndKey) < 2 ||
			strings.TrimSpace(tlsCertAndKey[0]) == "" ||
			strings.TrimSpace(tlsCertAndKey[1]) == "" {
			// there is not tls cert and key file info, just start http service
			ws.logger.Trace("webService", "Serve", "ListenAndServe for", webAddr)
			err = ws.server.ListenAndServe()
		} else {
			ws.logger.Trace("webService", "Serve", "ListenAndServeTLS for", webAddr)
			err = ws.server.ListenAndServeTLS(strings.TrimSpace(tlsCertAndKey[0]), strings.TrimSpace(tlsCertAndKey[1]))
		}

		if err != nil && err != http.ErrServerClosed {
			ws.logger.Error("webService", "Serve", "start web service", webAddr, "with error", err)
			panic(err)
		}

		ws.logger.Trace("webService", "Serve", "web service", webAddr, "is stopped", err)
	}()

}

func (ws *webService) Close() error {
	if ws.server == nil {
		return nil
	}

	// use context to control timeout of http.Server.Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := ws.server.Shutdown(ctx)
	if err != nil {
		ws.logger.Error("webService", "Close", "http server shutdown with error", err)
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

	ws.logger.Trace("webService", "dispatch", "service", ws.server.Addr,
		"get request from", remoteAddr, "path", r.URL.Path, "form", r.Form)

	if r.URL.Path == "/favicon.ico" {
		faviconHandler.ServeHTTP(w, r)
		ws.logger.Trace("webService", "dispatch", "service", ws.server.Addr,
			"handled request from", remoteAddr, "path", r.URL.Path)
		return
	}

	rsp := ws.checkAuth(r.URL.Path, r.RemoteAddr)
	if rsp != nil {
		ws.logger.Warn("webService", "dispatch", "service", ws.server.Addr, "checkAuth for",
			remoteAddr, "returned", rsp)
		ws.jsonResponseWithStatus(w, r, rsp, rsp.ErrNo)
		return
	}

	// threat others as interface access
	handler := ws.handerForPath(r.URL.Path)
	if handler == nil {
		ws.jsonResponseWithStatus(w, r, &ServiceResponse{
			ErrNo: StatusBadRequest,
			Err:   "bad request",
			Data:  map[int]int{},
		}, StatusBadRequest)
		ws.logger.Warn("webService", "dispatch", "service", ws.server.Addr, "invalid path", r.URL.Path,
			"remote address", remoteAddr)
		return
	}

	// we must read body first for part of data used AES and there might be a '=', ParseForm will wrong parse the data in that case
	body, _ := ioutil.ReadAll(r.Body)
	r.ParseForm()
	ws.logger.Debug("webService", "dispatch", "service", ws.server.Addr,
		"request url", r.RequestURI, "remote address", remoteAddr, "form", r.Form,
		"body", string(body))

	resp := handler(w, r, body)
	if resp != nil {
		if resp.StatusCode > 0 && resp.StatusCode != 200 {
			ws.jsonResponseWithStatus(w, r, resp, resp.StatusCode)
		} else {
			ws.jsonResponse(w, r, resp)
		}
	} else {
		ws.logger.Trace("webService", "dispatch", "service", ws.server.Addr, "remote address", remoteAddr,
			"path", r.URL.Path, "handler has responsed by itself")
	}
}

func (ws *webService) handerForPath(path string) RequestHandlerFunc {
	if v, ok := ws.handlers[path]; ok {
		return v
	}

	ws.logger.Warn("webService", "handerForPath", "service", ws.server.Addr, "not found handler for", path)
	return nil
}

func (ws webService) jsonResponse(w http.ResponseWriter, r *http.Request, data *ServiceResponse) {
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
		ws.logger.Error("webService", "jsonResponse", "service", ws.server.Addr,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with json marshal error", err, "response data", jsonRsp)
		fmt.Fprint(w, jsonRsp)
		return
	}

	_, err = fmt.Fprintf(w, "%v", string(dataBytes))
	if err != nil {
		ws.logger.Error("webService", "jsonResponse", "service", ws.server.Addr,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with error", err, "response data", data)
	} else {
		ws.logger.Trace("webService", "jsonResponse", "service", ws.server.Addr,
			"response request from", remoteAddr, "path", r.RequestURI,
			"with data", data)
	}
}

func (ws webService) jsonResponseWithStatus(w http.ResponseWriter, r *http.Request, data *ServiceResponse, status int) {
	remoteAddr := remoteAddrOfRequest(r)
	ws.logger.Trace("webService", "jsonResponseWithStatus", "service", ws.server.Addr,
		"response request from", remoteAddr, "path", r.RequestURI, "with status", status)
	w.WriteHeader(status)
	ws.jsonResponse(w, r, data)
}
