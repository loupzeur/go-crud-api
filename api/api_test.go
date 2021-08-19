package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/loupzeur/go-crud-api/middlewares"
	"github.com/loupzeur/go-crud-api/utils"
)

func TestGenRoute(t *testing.T) {
	//have a DB
	db, _ := gorm.Open("sqlite3", "gorm.db")
	SetDB(db)
	//init the router
	router := mux.NewRouter().StrictSlash(true)
	router.Use(middlewares.JwtAuthentication)

	//Create a default route for object test_object
	routes := CrudRoutes(&TestObject{},
		DefaultQueryAll, utils.NoRight, //QueryAll
		DefaultRightAccess, utils.NoRight, //Read
		DefaultRightAccess, utils.NoRight, //Create
		DefaultRightEdit, utils.NoRight, //Edit
		DefaultRightAccess, utils.NoRight, //Delete
	)

	//add routes to mdw
	middlewares.Routes = append(middlewares.Routes, routes...)

	for _, route := range middlewares.Routes {
		handler := route.HandlerFunc
		router.Methods(route.Method).Path(route.Pattern).Handler(handler).Name(route.Name)
	}

	//To test
	req, _ := http.NewRequest("GET", "/api/test_object", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Logf("Return Code : %d", rr.Code)

	if rr.Code != http.StatusOK {
		t.Errorf("Http api return error : %d", rr.Code)
	}

	//to execute in a main :
	//create srv and listen to it
	//srv := http.Server{
	//	Addr:    ":8080",
	//	Handler: router,
	//}
	//log.Fatal(srv.ListenAndServe())
}

//Test Object Route

type TestObject struct {
	Name string
}

//Default implement
func (c *TestObject) TableName() string {
	return "test_object" // can then be accessed in /api/test_object
}

//Validate to validate a model
func (c *TestObject) Validate() (map[string]interface{}, bool) {
	if c.Name == "" {
		return utils.Message(false, "Name is empty"), false
	}

	return nil, true
}

//OrderColumns return available columns
func (c *TestObject) OrderColumns() []string {
	return []string{}
}

//FilterColumns to return default columns to filter on
func (c *TestObject) FilterColumns() map[string]string {
	return map[string]string{}
}

//FindFromRequest to find Data from http request
func (c *TestObject) FindFromRequest(r *http.Request) error {
	return utils.DefaultFindFromRequest(r, GetDB(), c)
}

//QueryAllFromRequest to find Data from http request
func (c *TestObject) QueryAllFromRequest(r *http.Request, q *gorm.DB) *gorm.DB {
	return DefaultQueryAll(r, q)
}
