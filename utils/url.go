package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//Message to give a status and reason about errors or if data cannot be retrieved
func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

//Respond Create a JSON response with headers
func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

//RespondCode Create a JSON response with headers
func RespondCode(w http.ResponseWriter, data map[string]interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

//ReadInt return an int or a default value
func ReadInt(r *http.Request, param string, v int64) (int64, error) {
	p := r.FormValue(param)
	if p == "" {
		return v, nil
	}
	return strconv.ParseInt(p, 10, 64)
}

//ReadJSON read a json or return a response
func ReadJSON(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

//ReadIntURL read an URL parameter
func ReadIntURL(r *http.Request, param string) (int, error) {
	params := mux.Vars(r)
	value, err := strconv.Atoi(params[param])
	if err != nil {
		return 0, err
	}
	return value, err
}

//ReadIntURL read an URL parameter
func ReadFloatURL(r *http.Request, param string) (float64, error) {
	params := mux.Vars(r)
	value, err := strconv.ParseFloat(params[param], 64)
	if err != nil {
		return 0, err
	}
	return value, err
}

//ReadURL read an URL parameter
func ReadURL(r *http.Request, param string) string {
	return mux.Vars(r)[param]
}

//GetSha return a string of SHA from bytes
func GetSha(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
