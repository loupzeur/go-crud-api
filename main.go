package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/loupzeur/go-crud-api/api"
	"github.com/loupzeur/go-crud-api/middlewares"
	"github.com/loupzeur/go-crud-api/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rs/cors"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	gormopentracing "gorm.io/plugin/opentracing"
)

/*
https://github.com/jaegertracing/jaeger-client-go

*/

//basic HTTP server with jaeger debug
func main() {
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
	//tracer, closer := tracing.Init("crud-go-api")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	db.Use(gormopentracing.New())
	api.SetDB(db)
	db.AutoMigrate(&TestObject{})

	router := mux.NewRouter().StrictSlash(true)
	router.Use(middlewares.JaeggerLogger) //use of jaegger logger
	router.Use(middlewares.JwtAuthentication)

	middlewares.Routes = utils.Routes{utils.Route{Name: "test", Method: "GET", Pattern: "/test", Authorization: uint32(utils.NoRight), HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
		span := opentracing.GlobalTracer().StartSpan("simple-test")
		span.LogFields(log.String("test", "test"))
		span.Finish()
	}}}.Append(api.CrudRoutes(&TestObject{},
		utils.DefaultQueryAll, utils.NoRight, //QueryAll
		api.DefaultRightAccess, utils.NoRight, //Read
		api.DefaultRightAccess, utils.NoRight, //Create
		api.DefaultRightEdit, utils.NoRight, //Edit
		api.DefaultRightAccess, utils.NoRight, //Delete
	))

	for _, route := range middlewares.Routes {
		router.Methods(route.Method).Path(route.Pattern).Handler(route.HandlerFunc).Name(route.Name)
	}

	var handler http.Handler = router
	if os.Getenv("cors_dev") == "true" {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{
				"http://localhost:3000",
			},
			AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS", "OPTION"},
			AllowedHeaders:   []string{`Referer`, `user`, `Origin`, `DNT`, `User-Agent`, `X-Requested-With`, `If-Modified-Since`, `Cache-Control`, `Content-Type`, `Range`, `Authorization`, `X-Content-Length`, `X-Content-Name`, `X-Content-Extension`, `X-Chunk-Id`, `maxSize`, `maxWidth`, `maxHeight`, `fileType`},
			AllowCredentials: true,
		})

		handler = c.Handler(router)
	}
	srv := http.Server{
		ReadTimeout:       0,
		WriteTimeout:      600 * time.Second,
		IdleTimeout:       0,
		ReadHeaderTimeout: 0,
		Addr:              ":" + port,
		Handler:           handler,
	}
	srv.ListenAndServe()
}

//for example :
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
	return utils.DefaultFindFromRequest(r, api.GetDB(), c)
}

//QueryAllFromRequest to find Data from http request
func (c *TestObject) QueryAllFromRequest(r *http.Request, q *gorm.DB) *gorm.DB {
	return api.DefaultQueryAll(r, q)
}
