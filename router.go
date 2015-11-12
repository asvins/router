package router

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	routes           []*route
	baseInterceptors map[string][]Interceptor
}

//NewRouter = constructor for router
func NewRouter() *Router {
	return &Router{baseInterceptors: make(map[string][]Interceptor)}
}

//Interceptor is an inteface that objects that want to be used as interceptor for requests must implement.
type Interceptor interface {
	Intercept(rw http.ResponseWriter, r *http.Request) errors.Http
}

// Handler defines the prototype of the custom handlers for this router
type Handler func(http.ResponseWriter, *http.Request) errors.Http

// route struct has the route path, method handler e possible specific interceptors
type route struct {
	method       string
	regex        *regexp.Regexp
	reqParams    map[int]string
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
func (r *Router) Handle(pattern string, method string, handler Handler, interceptors []Interceptor) {
	switch method {
	case GET:
		r.doAddRoute(GET, pattern, handler, interceptors)
		break
	case PUT:
		r.doAddRoute(PUT, pattern, handler, interceptors)
		break
	case DELETE:
		r.doAddRoute(DELETE, pattern, handler, interceptors)
		break
	case POST:
		r.doAddRoute(POST, pattern, handler, interceptors)
		break
	}
}

//AddRoute adds a new route using path method, handler and a variadic number of interceptors
func (r *Router) AddRoute(pattern string, method string, handler http.HandlerFunc, interceptors ...Interceptor) {
	r.Handle(pattern, method, wrap(handler), interceptors)
}

//doAddRoute will add the specific route using method and string
func (r *Router) doAddRoute(method string, pattern string, handler Handler, interceptors []Interceptor) {
	if !strings.HasPrefix(pattern, "/") {
		fmt.Println("[ERROR] pattern should ALWAYS begin with '/'")
		panic("[ERROR] pattern should ALWAYS begin with '/'")
	}

	URISections := strings.Split(pattern, "/")

	j := 0
	reqParams := make(map[int]string)

	for i, section := range URISections {
		if strings.HasPrefix(section, ":") {
			// anything that has at least one char and it's not a '/'
			reqParams[j] = section
			URISections[i] = "([^/]+)"
			j++
		}
	}

	pattern = strings.Join(URISections, "/")
	reg, err := regexp.Compile(pattern)

	if err != nil {
		fmt.Println("[ERROR] Unable to add requested route: ", err)
		panic(err)
	}

	route := &route{}
	route.method = method
	route.regex = reg
	route.reqParams = reqParams
	route.handler = handler
	route.interceptors = interceptors

	r.routes = append(r.routes, route)
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
	requestURL := rq.URL.Path

	for _, route := range r.routes {

		// check method
		if route.method != rq.Method {
			continue
		}

		// check if regex match the request URL
		if !route.regex.MatchString(requestURL) {
			continue
		}

		// Example of FindStringSubmatch return:
		//	regex:	/api/user/([^/]+)/details/([^/]+)
		//	entered url:	/api/user/1234/details/12
		//	return:	[/api/user/1234/details/12 1234 12]
		matches := route.regex.FindStringSubmatch(requestURL)

		// this if is needed otherwise any url like '/api/users' would match with '/' if a route like that is registered
		if matches[0] != requestURL {
			continue
		}

		// put the params on url values to be able to access it from interceptors and handlers
		if len(route.reqParams) > 0 {
			values := rq.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.reqParams[i][1:], match)
			}

			rq.URL.RawQuery = url.Values(values).Encode()
		}

		// base interceptor execution
		err = r.executeBaseInterceptors(rq.URL.Path, w, rq) //base path interceptors
		if writeError(err, w) {
			return
		}

		// router interceptors execution
		err = route.executeInterceptors(w, rq) // route specific interceptors
		if writeError(err, w) {
			return
		}

		// handler execution
		err = route.handler(w, rq) // route handler
		if writeError(err, w) {
			return
		}

		return
	}

	// otherwise, serve static files? =s
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
