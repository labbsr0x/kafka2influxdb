package utils

import (
	"net/http"
)

type ServiceError struct {
	Not_Found bool
	Conflict  bool
	Internal  bool
	Forbidden bool
	Invalid   bool
}

func (r *ServiceError) SetStatusCode() int {
	if r.Not_Found {
		return http.StatusNotFound
	} else if r.Conflict {
		return http.StatusConflict
	} else if r.Internal {
		return http.StatusInternalServerError
	} else if r.Forbidden {
		return http.StatusForbidden
	} else if r.Invalid {
		return http.StatusBadRequest
	} else {
		return http.StatusOK
	}
}

func (r *ServiceError) Ok() bool {
	return !(r.Not_Found || r.Conflict || r.Internal || r.Forbidden || r.Invalid)
}
