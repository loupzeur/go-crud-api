package api

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

//Validation interface to validate stuff
type Validation interface {
	TableName() string
	Validate() (map[string]interface{}, bool)
	OrderColumns() []string
	FilterColumns() map[string]string
	FindFromRequest(r *http.Request) error
	QueryAllFromRequest(r *http.Request, q *gorm.DB) *gorm.DB
}

//Authed implement an element to set the id
type Authed interface {
	SetUserEmitter(userID uint)
}

//HistoryAble to store history on implementing objects
type HistoryAble interface {
	GetHistoryFields() map[string]string
	SetHistory([]map[string]interface{})
}
