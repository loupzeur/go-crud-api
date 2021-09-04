package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/loupzeur/go-crud-api/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

func TestCrash(t *testing.T) {
	defer setTracing().Close()
	router := mux.NewRouter().StrictSlash(true)
	router.Use(JaeggerLogger)
	router.Use(RecoverWrap)
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

func TestNoAuth(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(JwtAuthentication)
	router.Methods("GET").Path("/auth").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Name("auth")

	Routes = utils.Routes{{Name: "auth", Pattern: "/auth", Authorization: uint32(utils.NoRight)}}

	req, _ := http.NewRequest("GET", "/auth", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Logf("Return Code : %d", rr.Code)

	if rr.Code != http.StatusOK {
		t.Errorf("Http api return error : %d", rr.Code)
	}
}

func TestAuthOk(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(JwtAuthentication)
	router.Methods("GET").Path("/auth").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Name("auth")

	Routes = utils.Routes{{Name: "auth", Pattern: "/auth", Authorization: 1}} //require right 1

	req, _ := http.NewRequest("GET", "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+genToken(1, 1)) //give right 1
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Logf("Return Code : %d", rr.Code)

	if rr.Code != http.StatusOK {
		t.Errorf("Http api return error : %d", rr.Code)
	}
}

func TestAuthNOk(t *testing.T) {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(JwtAuthentication)
	router.Methods("GET").Path("/auth").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Name("auth")

	Routes = utils.Routes{{Name: "auth", Pattern: "/auth", Authorization: 2}} //the token right require 2

	req, _ := http.NewRequest("GET", "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+genToken(1, 1)) //token only give right 1 to user 1
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	t.Logf("Return Code : %d", rr.Code)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Http api return error : %d", rr.Code)
	}
}

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

func genToken(userId uint, right utils.RightBits) string {
	tk := &utils.Token{
		UserId:     userId,
		UserRights: right,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	return tokenString
}
