package api

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

//DefaultQueryAll default request for GetAll
func DefaultQueryAll(r *http.Request, req *gorm.DB) *gorm.DB {
	return req
}

//DefaultRightAccess return true right handler
func DefaultRightAccess(r *http.Request, data interface{}) bool {
	return true
}

//DefaultRightEdit return true right handler
func DefaultRightEdit(r *http.Request, data interface{}, data2 interface{}) bool {
	return true
}
