package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
)

func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-msgpack")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		err := MethodNotAllow()
		w.WriteHeader(err.Code)
		res, _ := msgpack.Marshal(err)
		w.Write(res)
	})
}

func notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-msgpack")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		err := NotFound()
		w.WriteHeader(err.Code)
		res, _ := msgpack.Marshal(err)
		w.Write(res)
	})
}

func HandleRequest(w http.ResponseWriter, r *http.Request, f func(r2 *http.Request) (interface{}, HttpError)) {
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	data, err := f(r)
	var res []byte
	if err.Message != "" {
		w.WriteHeader(err.Code)
		res, _ = msgpack.Marshal(err)
	} else {
		res, _ = msgpack.Marshal(data)
	}

	fmt.Println(res)
	w.Write(res)
}

func GenerateMuxRouter (routes []Route, middleware []mux.MiddlewareFunc) *mux.Router {
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = methodNotAllowedHandler()
	r.NotFoundHandler = notFoundHandler()

	for i, _ := range routes {
		rr := GetRouteAt(i)
		r.HandleFunc(rr.Url, func(w http.ResponseWriter, r *http.Request) {
			HandleRequest(w, r, rr.Func)
		}).Methods(rr.HttpMethod).Name(rr.Name)
	}

	r.Use(ContentTypeMiddleware)
	for _, mwr := range middleware {
		mw := mwr
		r.Use(mw)
	}

	return r
}

func StartHttpServer(port string) error {
	r := GenerateMuxRouter(routes, middlewares)
	return http.ListenAndServe(port, r)
}