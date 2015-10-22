package router

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

/*
*	This router implements only GET POST and PUT methods.
* interceptors can be added to specific routes or to base paths
 */

//HTTP METHODS
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

//Interceptor is an inteface that objects that want to be used as interceptor for requests must implement.
type Interceptor interface {
	Intercept(rw http.ResponseWriter, r *http.Request) error
}

// route struct has the route path, method handler e possible specific interceptors
type route struct {
	path         string
	method       string
	handler      http.HandlerFunc
	interceptors []Interceptor
}

// if an error occurs, the interceptor chain will stop immediately
func (r route) executeInterceptors(w http.ResponseWriter, rq *http.Request) error {
	var err error
	for _, interceptor := range r.interceptors {
		err = interceptor.Intercept(w, rq)
		if err != nil {
			return err
		}
	}
	return nil
}

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

//AddBaseInterceptor adds a new interceptor to a base path of a route
//The ideia is that, for example, all requests on /api/.... have a specific interceptor(eg: auth)
func (r *Router) AddBaseInterceptor(path string, interceptor Interceptor) {
	r.baseInterceptors[path] = append(r.baseInterceptors[path], interceptor)
}

//AddRoute adds a new route using path method, handler and a variadic number of interceptors
func (r *Router) AddRoute(path string, method string, handler http.HandlerFunc, interceptors ...Interceptor) {
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

//doAddRoute will add the specific route using method and string
func (r *Router) doAddRoute(method string, path string, handler http.HandlerFunc, interceptors []Interceptor) {
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
func (r *Router) executeBaseInterceptors(path string, w http.ResponseWriter, rq *http.Request) error {
	subpaths := strings.Split(path, "/")
	var err error
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

// ServeHTTP Implements interface http.Handler
// It will behave like this:
//	i) base interceptors execution
//	ii) route specific interceptor execution
//  iii) route handler execution
//
//	If any of the interceptors returns an error, the interceptor chain will be stopped immediately
func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	route := r.routes[rq.URL.Path+rq.Method]
	if route != nil {
		err := r.executeBaseInterceptors(rq.URL.Path, w, rq) //base path interceptors
		if err != nil {
			return
		}
		err = route.executeInterceptors(w, rq) // route specific interceptors
		if err != nil {
			return
		}
		route.handler(w, rq) // route handler
		return
	}

	// think again about static files...
	match, err := regexp.MatchString(servingFileRegex, rq.URL.Path)

	if err != nil {
		//Shouldn't get here...
		log.Fatal(err)
	}

	if match {
		http.ServeFile(w, rq, rq.URL.Path[1:])
		return
	}

	http.NotFound(w, rq)
}
