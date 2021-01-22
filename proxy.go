package webservice

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type proxy struct {
	logger Logger
}

func (p proxy) jsonResponse(w http.ResponseWriter, r *http.Request, data *ServiceResponse) {
	jsonResponse(w, r, data, "http proxy", p.logger)
}

func (p proxy) ForwardRequest(w http.ResponseWriter, r *http.Request, target *url.URL) {
	p.logger.Trace("entered...")
	defer p.logger.Trace("done.")

	resp, err := p.agentRequest(r, target)
	if err != nil {
		p.logger.Trace("agent request failed with", err)
		p.jsonResponse(w, r, &ServiceResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
			Data:    map[int]int{},
		})
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	written, err := io.Copy(w, resp.Body)
	p.logger.Trace("has written", written, "bytes to", r.RemoteAddr)
}

func (p proxy) agentRequest(r *http.Request, target *url.URL) (*http.Response, error) {
	if target == nil || r == nil {
		p.logger.Warn("invalid arument", "target", target, "request", r)
		return nil, ErrorInvalidArgument
	}

	tr := &http.Transport{
		Proxy: func(_ *http.Request) (*url.URL, error) { return target, nil },
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Minute,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Minute,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		DisableCompression:    true,
	}

	c := &http.Client{Transport: tr}

	req, err := http.NewRequest(r.Method, target.String(), r.Body)
	if err != nil {
		p.logger.Trace("new request failed with", err)
		return nil, err
	}
	for k, v := range r.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	return c.Do(req)
}

func (p proxy) AgentRequest(r *http.Request, target *url.URL) (*http.Response, error) {
	p.logger.Trace("entered...")
	defer p.logger.Trace("done.")

	return p.agentRequest(r, target)
}
