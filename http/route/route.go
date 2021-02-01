package route

import (
	http2 "github.com/dimall-id/lumos/http"
	"net/http"
)

type Route struct {
	Name string
	HttpMethod string
	Url	string
	Func func (r *http.Request) (interface{}, http2.HttpError)
}

func (r *Route) IsValid() bool {
	if r.Name == "" || r.HttpMethod == "" || r.Url == "" || r.Func == nil {
		return false
	}
	return true
}

func (r *Route) Equal (r2 Route) bool {
	if r.HttpMethod == r2.HttpMethod && r.Url == r2.Url {
		return true
	}
	return false
}

var routes []Route

func isExist (route Route) (bool, int) {
	for i, r := range routes {
		if r.Equal(route) {
			return true,i
		}
	}
	return false,-1
}

func Add (route Route) error {
	if oke,_ := isExist(route); oke {
		return &ExistingRouteError{route: route}
	} else if route.IsValid() {
		return &InvalidRouteError{route: route}
	} else {
		routes = append(routes, route)
		return nil
	}
}

func AddAll (rs []Route) error {
	for _,route := range rs {
		if oke,_ := isExist(route); oke {
			return &ExistingRouteError{route: route}
		} else if route.IsValid() {
			return &InvalidRouteError{route: route}
		} else {
			routes = append(routes, route)
		}
	}
	return nil
}

func Get (method string, url string) Route {
	route := Route{
		HttpMethod: method,
		Url: url,
		Func: nil,
	}
	if oke, ind := isExist(route); oke {
		return routes[ind]
	}
	return Route{}
}

func GetAt (i int) Route {
	return routes[i]
}

func GetAll () []Route {
	return routes
}