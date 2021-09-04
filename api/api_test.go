package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/loupzeur/go-crud-api/middlewares"
	"github.com/loupzeur/go-crud-api/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setTracing() io.Closer {
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	cfg := config.Configuration{
		ServiceName: "crud-go-api",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, _ := cfg.NewTracer(
		config.Logger(jaeger.StdLogger), config.ZipkinSharedRPCSpan(true),
		config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		config.Extractor(opentracing.HTTPHeaders, zipkinPropagator))
	opentracing.SetGlobalTracer(tracer)
	return closer
}

func TestCrash(t *testing.T) {
	defer setTracing().Close()
	router := mux.NewRouter().StrictSlash(true)
	router.Use(middlewares.JaeggerLogger)
	router.Use(middlewares.RecoverWrap)
	router.Methods("GET").Path("/crash").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("mouhahahahaha a error occured")
	}).Name("crash")

	req, _ := http.NewRequest("GET", "/crash", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Logf("Return Code : %d", rr.Code)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Http api return error : %d", rr.Code)
	}
}
func TestGenRoute(t *testing.T) {
	//have a DB
	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
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
