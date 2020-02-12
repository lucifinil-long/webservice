package webservice

import (
	"strings"
)

func (ws webService) checkAuth(path, remoteAddr string) *ServiceResponse {
	parts := strings.Split(remoteAddr, ":")
	if len(parts) != 2 {
		return &ServiceResponse{
			ErrNo: StatusBadRequest,
			Err:   "invalid remoteaddr",
			Data:  map[int]int{},
		}
	}

	auth := ws.haveAuth(path, parts[0])
	ws.logger.Trace("webService", "checkAuth", path, parts[0], auth)
	if !auth {
		return &ServiceResponse{
			ErrNo: StatusForbidden,
			Err:   "have perm limit, and your ip not in whitelist",
			Data:  map[int]int{},
		}
	}

	return nil
}

func (ws webService) haveAuth(path, ip string) bool {
	if len(ws.authMap) == 0 {
		return true
	}

	if v, ok := ws.authMap[path]; ok {
		_, ok = v[ip]
		return ok
	}

	return true
}
