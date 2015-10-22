package router

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

/*
*	Common variables
 */
var (
	server                       *httptest.Server
	endpointURL                  string
	indexInterceptorCount        int = 2
	indexInterceptorChan             = make(chan int, indexInterceptorCount)
	apiInterceptorCount          int = 3
	apiInterceptorChan               = make(chan int, apiInterceptorCount)
	apiUsersNameInterceptorCount int = 1
	apiUsersNameInterceptorChan      = make(chan int, apiUsersNameInterceptorCount)
	failInterceptorCount         int = 2
	failInterceptorChan              = make(chan int, failInterceptorCount)
)

/*
*	Base interceptors
 */
type indexInterceptor struct{}

func (t indexInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[indexInterceptor]")
	indexInterceptorChan <- 1
	return nil
}

type indexInterceptor2 struct{}

func (t indexInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[indexInterceptor2]")
	indexInterceptorChan <- 1
	return nil
}

type apiInterceptor struct{}

func (t apiInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[apiInterceptor]")
	apiInterceptorChan <- 1
	return nil
}

type apiInterceptor2 struct{}

func (t apiInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[apiInterceptor2]")
	apiInterceptorChan <- 1
	return nil
}

type apiInterceptor3 struct{}

func (t apiInterceptor3) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[apiInterceptor3]")
	apiInterceptorChan <- 1
	return nil
}

type failInterceptor struct{}

func (t failInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[failInterceptor] - will return an error!")
	err := errors.New("failInterceptor error returned")
	fmt.Fprint(rw, err.Error())
	failInterceptorChan <- 1
	return err
}

type failInterceptor2 struct{}

func (t failInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[failInterceptor2] - should not be called")
	failInterceptorChan <- 1
	return nil
}

/*
*	Specific routes interceptors
 */

type apiUsersNameInterceptor struct{}

func (t apiUsersNameInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[apiUsersNameInterceptor]")
	apiUsersNameInterceptorChan <- 1
	return nil
}

/*
*	Init function. Will create a mocked server for testing
 */
func init() {
	r := NewRouter()

	r.AddRoute("/", GET, func(w http.ResponseWriter, apiRouter *http.Request) {
		fmt.Fprint(w, "Request made to '/'")
	}, &indexInterceptor{}, &indexInterceptor2{})

	r.AddRoute("/api/users", GET, func(w http.ResponseWriter, apiRouter *http.Request) {
		fmt.Fprint(w, "Request made to '/api/users'")
	})

	r.AddRoute("/api/users/name", GET, func(w http.ResponseWriter, apiRouter *http.Request) {
		fmt.Fprint(w, "Request made to '/api/users/name'")
	})

	r.AddRoute("/willfail/now", GET, func(w http.ResponseWriter, apiRouter *http.Request) {
		fmt.Fprint(w, "Request made to '/willfail/now'")
	})

	r.AddBaseInterceptor("/api", &apiInterceptor{})
	r.AddBaseInterceptor("/api", &apiInterceptor2{})
	r.AddBaseInterceptor("/api", &apiInterceptor3{})
	r.AddBaseInterceptor("/api/users/name", &apiUsersNameInterceptor{})
	r.AddBaseInterceptor("/willfail/now", &failInterceptor{})
	r.AddBaseInterceptor("/willfail/now", &failInterceptor2{})

	server = httptest.NewServer(r)
	endpointURL = server.URL
}

func get(path string) (*http.Response, error) {
	reader := strings.NewReader(``)
	request, err := http.NewRequest("GET", endpointURL+path, reader)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func TestSpecificRouteInterceptor(t *testing.T) {
	fmt.Println("-- TestSpecificRouteInterceptor start --")
	response, err := get("")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	interceptorCount := 0
	for i := 0; i < indexInterceptorCount; i++ {
		interceptorCount += <-indexInterceptorChan
	}

	if interceptorCount != indexInterceptorCount {
		t.Error("Not all interceptors called for '/'")
	}
	fmt.Println("-- TestSpecificRouteInterceptor end --\n")
}

func TestBaseRouteInterceptor(t *testing.T) {
	fmt.Println("-- TestBaseRouteInterceptor start --")
	response, err := get("/api/users")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	interceptorCount := 0
	for i := 0; i < apiInterceptorCount; i++ {
		interceptorCount += <-apiInterceptorChan
	}

	if interceptorCount != apiInterceptorCount {
		t.Error("Not all interceptors called for '/api/users'")
	}
	fmt.Println("-- TestBaseRouteInterceptor end --\n")
}

func TestBaseAndSpecificInterceptor(t *testing.T) {
	fmt.Println("-- TestBaseRouteInterceptor start --")
	response, err := get("/api/users/name")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	interceptorCount := 0
	for i := 0; i < apiInterceptorCount; i++ {
		interceptorCount += <-apiInterceptorChan
	}

	for i := 0; i < apiUsersNameInterceptorCount; i++ {
		interceptorCount += <-apiUsersNameInterceptorChan
	}

	if interceptorCount != (apiInterceptorCount + apiUsersNameInterceptorCount) {
		t.Error("Not all interceptors called for '/api/users/name'")
	}
	fmt.Println("-- TestBaseRouteInterceptor end --\n")
}

func TestBaseRouteError(t *testing.T) {
	fmt.Println("-- TestBaseRouteError start --")
	response, err := get("/willfail/now")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	// first timeout - should NOT occur
	select {
	case <-failInterceptorChan:
		fmt.Println("First failInterceptor ran - OK")
	case <-time.After(time.Second * 1):
		t.Error("Timeout occured when receiving first info from failInterceptorChan")
	}

	// this timeout MUST occur
	select {
	case <-failInterceptorChan:
		fmt.Println("Second failInterceptor ran - ERROR")
		t.Error("Interceptor error didn't break the interceptor chain")
	case <-time.After(time.Second * 1):
		fmt.Println("Second failInterceptor timeout - OK")
	}

	fmt.Println("-- TestBaseRouteError end --\n")
}