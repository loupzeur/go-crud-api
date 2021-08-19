package utils

import "net/http"

//Route define a route with url and rights required to access it
type Route struct {
	Name          string           `json:"name"`
	Method        string           `json:"method"`
	Pattern       string           `json:"pattern"`
	HandlerFunc   http.HandlerFunc `json:"-"`
	Authorization uint32           `json:"auth"`
}

//Routes an array of route
type Routes []Route

//Append Add routes to Routes
func (r Routes) Append(routes []Route) Routes {
	r = append(r, routes...)
	return r
}

//Get Return route by name
func (r Routes) Get(name string) Route {
	ret := Route{}
	for _, v := range r {
		if v.Name == name {
			return v
		}
	}
	return ret
}
