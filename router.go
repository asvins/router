package router

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/asvins/router/errors"
	"github.com/unrolled/render"
)

/*
*	This router implements only GET POST and PUT methods.
* interceptors can be added to specific routes or to base paths
 */
var rend *render.Render

func init() {
	rend = render.New()
}

//HTTP METHODS
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

const (
	servingFileRegex = "^.*\\.(html|css|js)$"
)

//Router struct containing the routes and base interceptors
type Router struct {
	routes           map[string]*route
	baseInterceptors map[string][]Interceptor
}

//NewRouter = constructor for router
func NewRouter() *Router {
	return &Router{routes: make(map[string]*route), baseInterceptors: make(map[string][]Interceptor)}
}

//Interceptor is an inteface that objects that want to be used as interceptor for requests must implement.
type Interceptor interface {
	Intercept(rw http.ResponseWriter, r *http.Request) errors.Http
}

// Handler defines the prototype of the custom handlers for this router
type Handler func(http.ResponseWriter, *http.Request) errors.Http

// route struct has the route path, method handler e possible specific interceptors
type route struct {
	path         string
	method       string
	handler      Handler
	interceptors []Interceptor
}

// wrap converts a http.handler into a router.Handler
func wrap(handler http.HandlerFunc) Handler {
	return Handler(func(rw http.ResponseWriter, r *http.Request) errors.Http {
		handler(rw, r)
		return nil
	})
}

// if an error occurs, the interceptor chain will stop immediately
func (r route) executeInterceptors(w http.ResponseWriter, rq *http.Request) errors.Http {
	var err errors.Http
	for _, interceptor := range r.interceptors {
		err = interceptor.Intercept(w, rq)
		if err != nil {
			return err
		}
	}
	return nil
}

//AddBaseInterceptor adds a new interceptor to a base path of a route
//The ideia is that, for example, all requests on /api/.... have a specific interceptor(eg: auth)
func (r *Router) AddBaseInterceptor(path string, interceptor Interceptor) {
	r.baseInterceptors[path] = append(r.baseInterceptors[path], interceptor)
}

// Handle adds a new route with  router.Handler as handler
// If you choose to use this method, DON'T WRITE INTO THE RESPONSE WRITER IF YOU RETURN AN ERROR
//	if you Return a router.error.Http, the router will automatically return the error as a json on the response
func (r *Router) Handle(path string, method string, handler Handler, interceptors []Interceptor) {
	switch method {
	case GET:
		r.doAddRoute(GET, path, handler, interceptors)
		break
	case PUT:
		r.doAddRoute(PUT, path, handler, interceptors)
		break
	case DELETE:
		r.doAddRoute(DELETE, path, handler, interceptors)
		break
	case POST:
		r.doAddRoute(POST, path, handler, interceptors)
		break
	}
}

//AddRoute adds a new route using path method, handler and a variadic number of interceptors
func (r *Router) AddRoute(path string, method string, handler http.HandlerFunc, interceptors ...Interceptor) {
	r.Handle(path, method, wrap(handler), interceptors)
}

//doAddRoute will add the specific route using method and string
func (r *Router) doAddRoute(method string, path string, handler Handler, interceptors []Interceptor) {
	if r.routes[path+method] != nil {
		fmt.Printf("route with path '%s' with method '%s' already added. The second one will be ignored", path, method)
		return
	}

	route := &route{}
	route.path = path
	route.method = method
	route.handler = handler
	route.interceptors = interceptors

	r.routes[path+method] = route
}

// executeBaseInterceptors executes all interceptors in a given path.
// ex: request on /api/consumer/info
//  ->will execute base interceptors for:
//		i)'/'
//		ii)'/api'
//		iii)'/api/consumer'
//		iv)'/api/consumer/info'
func (r *Router) executeBaseInterceptors(path string, w http.ResponseWriter, rq *http.Request) errors.Http {
	subpaths := strings.Split(path, "/")
	var err errors.Http
	currPath := "/"

	for i := 1; i <= len(subpaths); i++ {
		for _, interceptor := range r.baseInterceptors[currPath] {
			err = interceptor.Intercept(w, rq)
			if err != nil {
				return err
			}
		}
		if i == len(subpaths) || subpaths[i] == "" {
			break
		}
		if i == 1 {
			currPath += subpaths[i]
		} else {
			currPath += "/" + subpaths[i]
		}
	}

	return nil
}

// writeError writes the errors.Http into a JSON with the correct status code.
// Return:
//	- true if did wrote an error(err argument != nil)
//	- false if didn't
func writeError(err errors.Http, w http.ResponseWriter) bool {
	if err != nil {
		rend.JSON(w, err.Code(), err)
		return true
	}
	return false
}

// ServeHTTP Implements interface http.Handler
// It will behave like this:
//	i) base interceptors execution
//	ii) route specific interceptor execution
//  iii) route handler execution
//
//	If any of the interceptors returns an error, the interceptor chain will be stopped immediately
func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	var err errors.Http

	route := r.routes[rq.URL.Path+rq.Method]
	if route != nil {
		err = r.executeBaseInterceptors(rq.URL.Path, w, rq) //base path interceptors
		if writeError(err, w) {
			return
		}

		err = route.executeInterceptors(w, rq) // route specific interceptors

		if writeError(err, w) {
			return
		}

		err = route.handler(w, rq) // route handler
		if writeError(err, w) {
			return
		}

		return
	}

	// think again about static files...
	match, matchErr := regexp.MatchString(servingFileRegex, rq.URL.Path)

	if matchErr != nil {
		log.Fatal(matchErr)
	}

	if match {
		http.ServeFile(w, rq, rq.URL.Path[1:])
		return
	}

	http.NotFound(w, rq)
}
