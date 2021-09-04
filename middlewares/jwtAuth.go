package middlewares

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/loupzeur/go-crud-api/utils"
	"github.com/opentracing/opentracing-go"

	jwt "github.com/dgrijalva/jwt-go"
)

//Routes from main
var Routes utils.Routes

//JwtAuthentication jwt auth checker handler -> set token for user
func JwtAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curRoute := mux.CurrentRoute(r)
		curRouter := utils.Route{}
		userRight := uint32(utils.NoRight) //default to no right to check for no auth routes
		defer func() {                     //simple logging
			auth, ctx := opentracing.StartSpanFromContext(r.Context(), "auth")
			r = r.WithContext(ctx)
			defer auth.Finish()
			auth.LogKV("url", r.URL.String())
			log.Println(r.URL.String(), r.UserAgent(), r.Header, w.Header(), curRoute.GetName())
		}()
		for _, value := range Routes {
			if value.Name == curRoute.GetName() {
				curRouter = value
				if (userRight & value.Authorization) == value.Authorization {
					next.ServeHTTP(w, r)
					return
				}
				break
			}

		}

		var response map[string]interface{}
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			response = utils.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			response = utils.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		tokenPart := splitted[1] //Grab the token part, what we are truly interested in
		tk := &utils.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil { //Malformed token, returns with http code 403 as usual
			response = utils.Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		if !token.Valid { //Token is invalid, maybe not signed on this server
			response = utils.Message(false, "Token is not valid.")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		if uint32(tk.UserRights)&curRouter.Authorization != curRouter.Authorization { //Token is invalid, maybe not signed on this server
			response = utils.Message(false, "Authorization required")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
		ctx := context.WithValue(r.Context(), "user", *tk)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

//GetAllRoutes return all routes in the app
func GetAllRoutes(w http.ResponseWriter, r *http.Request) {
	msg := utils.Message(true, "All Routes")
	if os.Getenv("ENV") != "PROD" {
		msg["data"] = Routes
	}
	utils.Respond(w, msg)
}
