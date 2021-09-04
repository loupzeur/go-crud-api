package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/loupzeur/go-crud-api/middlewares"
	"github.com/loupzeur/go-crud-api/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rs/cors"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

/*
https://github.com/jaegertracing/jaeger-client-go

*/

//basic HTTP server with jaeger debug
func main() {
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, _ := cfg.New("crud-go-api",
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

	router := mux.NewRouter().StrictSlash(true)
	router.Use(middlewares.JaeggerLogger)
	router.Use(middlewares.JwtAuthentication)

	middlewares.Routes = utils.Routes{utils.Route{Name: "test", Method: "GET", Pattern: "/test", Authorization: uint32(utils.NoRight), HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
		span := opentracing.GlobalTracer().StartSpan("simple-test")
		span.LogFields(log.String("test", "test"))
		span.Finish()
	}}}

	for _, route := range middlewares.Routes {
		router.HandleFunc(route.Pattern, route.HandlerFunc)
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
