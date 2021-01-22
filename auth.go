package webservice

import (
	"net/http"
	"strings"
)

func (ws webService) checkAuth(path, remoteAddr string) *ServiceResponse {
	parts := strings.Split(remoteAddr, ":")
	if len(parts) != 2 {
		return &ServiceResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid remoteaddr",
			Data:    map[int]int{},
		}
	}

	auth := ws.haveAuth(path, parts[0])
	ws.Logger.Trace("webService", "checkAuth", path, parts[0], auth)
	if !auth {
		return &ServiceResponse{
			Status:  http.StatusForbidden,
			Message: "have perm limit, and your ip not in whitelist",
			Data:    map[int]int{},
		}
	}

	return nil
}

func (ws webService) haveAuth(path, ip string) bool {
	if len(ws.AuthMap) == 0 {
		return true
	}

	if v, ok := ws.AuthMap[path]; ok {
		_, ok = v[ip]
		return ok
	}

	return true
}
