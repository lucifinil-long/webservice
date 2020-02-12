package webservice

const (
	// ErrSuccess indicates success
	ErrSuccess = 0
	// StatusOK indicates http OK status
	StatusOK = 200
	// StatusBadRequest indicates http bad request status
	StatusBadRequest = 400
	// StatusLackPara indicates http lack parameter status
	StatusLackPara = 401
	// StatusForbidden indicates http forbidden status
	StatusForbidden = 403
	// StatusNotFound indicates http not found status
	StatusNotFound = 404
	// StatusInternal indicates http server internal status
	StatusInternal = 500
	// StatusDBError indicates http DB error status
	StatusDBError = 501
)

// ServiceResponse defines union web service response
type ServiceResponse struct {
	ErrNo      int         `json:"errno"`
	Err        string      `json:"errmsg"`
	Data       interface{} `json:"data"`
	Indent     bool        `json:"-"`
	StatusCode int         `json:"-"`
}
