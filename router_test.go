package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	routerErrors "github.com/asvins/router/errors"
)

/*
*	Common variables
 */
var (
	r                            *Router
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

func (t indexInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[indexInterceptor]")
	indexInterceptorChan <- 1
	return nil
}

type indexInterceptor2 struct{}

func (t indexInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[indexInterceptor2]")
	indexInterceptorChan <- 1
	return nil
}

type apiInterceptor struct{}

func (t apiInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[apiInterceptor]")
	apiInterceptorChan <- 1
	return nil
}

type apiInterceptor2 struct{}

func (t apiInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[apiInterceptor2]")
	apiInterceptorChan <- 1
	return nil
}

type apiInterceptor3 struct{}

func (t apiInterceptor3) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[apiInterceptor3]")
	apiInterceptorChan <- 1
	return nil
}

type failInterceptor struct{}

func (t failInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[failInterceptor] - will return an error!")
	err := routerErrors.BadRequest("Bad request MOCK")
	failInterceptorChan <- 1
	return err
}

type failInterceptor2 struct{}

func (t failInterceptor2) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[failInterceptor2] - should not be called")
	failInterceptorChan <- 1
	return nil
}

/*
*	Specific routes interceptors
 */

type apiUsersNameInterceptor struct{}

func (t apiUsersNameInterceptor) Intercept(rw http.ResponseWriter, r *http.Request) routerErrors.Http {
	fmt.Println("[apiUsersNameInterceptor]")
	apiUsersNameInterceptorChan <- 1
	return nil
}

/*
*	Init function. Will create a mocked server for testing
 */
func init() {
	r = NewRouter()

	r.AddRoute("/", GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/'")
	}, &indexInterceptor{}, &indexInterceptor2{})

	r.AddRoute("/api/users", GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/api/users'")
	})

	r.AddRoute("/api/users/name", GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/api/users/name'")
	})

	r.AddRoute("/willfail/now", GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/willfail/now'")
	})

	r.Handle("/handler/unauthorized", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		return routerErrors.Unauthorized("You shall not pass")
	}, []Interceptor{})

	r.Handle("/handler/badrequest", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		return routerErrors.BadRequest("That's a bad request")
	}, []Interceptor{})

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
	request, err := http.NewRequest(GET, endpointURL+path, reader)
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
	fmt.Println("-- TestBaseAndSpecificInterceptor start --")
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

func TestHandleUnauthorized(t *testing.T) {
	fmt.Println("-- TestHandleUnauthorized start --")
	response, err := get("/handler/unauthorized")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusUnauthorized {
		t.Error("Status Code should be", http.StatusUnauthorized, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleUnauthorized end --\n")
}

func TestHandleBadRequest(t *testing.T) {
	fmt.Println("-- TestHandleBadRequest start --")
	response, err := get("/handler/badrequest")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusBadRequest {
		t.Error("Status Code should be", http.StatusBadRequest, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleBadRequest end --\n")
}

func TestHandleResourceRoute(t *testing.T) {
	fmt.Println("-- TestHandleResourceRoute start --")

	// Adding the route definition
	r.Handle("/user/:uid", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		fmt.Println("Rquest made to '/user/:uid'")
		params := rq.URL.Query()

		uid := params.Get("uid")
		fmt.Println("Uid = ", uid)

		uid_int, err := strconv.Atoi(uid)
		if err != nil {
			fmt.Println("[ERROR] :uid not an integer value!")
			t.Error(err)
		}

		if uid_int != 1234 {
			fmt.Println("Expected uid: 1234 Got ", uid_int)
			t.Error("Expected uid: 1234 Got ", uid_int)
		}

		return nil
	}, []Interceptor{})

	response, err := get("/user/1234")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusOK {
		t.Error("Status Code should be", http.StatusOK, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleResourceRoute end --\n")
}

func TestHandleResouceWithQueryString(t *testing.T) {
	fmt.Println("-- TestHandleResouceWithQueryString start --")

	// Adding the route definition
	r.Handle("/user/:uid/details/:did", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		fmt.Println("Rquest made to '/user/:uid/details/:did'")
		params := rq.URL.Query()

		uid := params.Get("uid")
		fmt.Println("Uid = ", uid)
		did := params.Get("did")
		fmt.Println("Did = ", did)
		cid := params.Get("cid")
		fmt.Println("Cid = ", cid)

		uid_int, err := strconv.Atoi(uid)
		if err != nil {
			fmt.Println("[ERROR] :uid not an integer value!")
			t.Error(err)
		}

		if uid_int != 1234 {
			fmt.Println("Expected uid: 1234 Got ", uid_int)
			t.Error("Expected uid: 1234 Got ", uid_int)
		}

		did_int, err := strconv.Atoi(did)
		if err != nil {
			fmt.Println("[ERROR] :did not an integer value!")
			t.Error(err)
		}

		if did_int != 5678 {
			fmt.Println("Expected uid: 5678 Got ", did_int)
			t.Error("Expected uid: 5678 Got ", did_int)
		}

		cid_int, err := strconv.Atoi(cid)
		if err != nil {
			fmt.Println("[ERROR] :cid not an integer value!")
			t.Error(err)
		}

		if cid_int != 9012 {
			fmt.Println("Expected cid: 9012 Got ", cid_int)
			t.Error("Expected uid: 9012 Got ", cid_int)
		}

		return nil
	}, []Interceptor{})

	response, err := get("/user/1234/details/5678?cid=9012")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusOK {
		t.Error("Status Code should be", http.StatusOK, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleResouceWithQueryString end --\n")
}

func TestHandleResourceRouteNotInserted(t *testing.T) {
	fmt.Println("-- TestHandleResourceRouteNotInserted start --")

	response, err := get("/user/1234/details")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusNotFound {
		t.Error("Status Code should be", http.StatusNotFound, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleResourceRouteNotInserted end --\n")
}

func TestHandleResouceWithQueryString2(t *testing.T) {
	fmt.Println("-- TestHandleResouceWithQueryString2 start --")

	// Adding the route definition
	r.Handle("/user/:uid/details", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		fmt.Println("Rquest made to '/user/:uid/details'")
		params := rq.URL.Query()

		uid := params.Get("uid")
		fmt.Println("Uid = ", uid)
		did := params.Get("did")
		fmt.Println("Did = ", did)
		cid := params.Get("cid")
		fmt.Println("Cid = ", cid)

		uid_int, err := strconv.Atoi(uid)
		if err != nil {
			fmt.Println("[ERROR] :uid not an integer value!")
			t.Error(err)
		}

		if uid_int != 1234 {
			fmt.Println("Expected uid: 1234 Got ", uid_int)
			t.Error("Expected uid: 1234 Got ", uid_int)
		}

		did_int, err := strconv.Atoi(did)
		if err != nil {
			fmt.Println("[ERROR] :did not an integer value!")
			t.Error(err)
		}

		if did_int != 5678 {
			fmt.Println("Expected uid: 5678 Got ", did_int)
			t.Error("Expected uid: 5678 Got ", did_int)
		}

		cid_int, err := strconv.Atoi(cid)
		if err != nil {
			fmt.Println("[ERROR] :cid not an integer value!")
			t.Error(err)
		}

		if cid_int != 9012 {
			fmt.Println("Expected cid: 9012 Got ", cid_int)
			t.Error("Expected cid: 9012 Got ", cid_int)
		}

		return nil
	}, []Interceptor{})

	response, err := get("/user/1234/details?did=5678&cid=9012")
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	fmt.Println("StatusCode:", response.StatusCode)

	if response.StatusCode != http.StatusOK {
		t.Error("Status Code should be", http.StatusOK, " Got", response.StatusCode)
	}

	fmt.Println("-- TestHandleResouceWithQueryString2 end --\n")
}

func TestAddingMissConstructedRoute(t *testing.T) {
	fmt.Println("-- TestAddingMissConstructedRoute start --\n")

	defer func() {
		if recv := recover(); recv == nil {
			t.Error("[ERROR] Should have panicd because function tried to add a misscontructed route")
		} else {
			fmt.Println("-- TestAddingMissConstructedRoute end --\n")
		}
	}()

	// Trying to add route that do not begins with '/'
	fmt.Println("[INFO] Should print error from router.Handle for trying to add route that doesn't begin with '/'")
	r.Handle("user/:uid/details", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		fmt.Println("Rquest made to 'user/:uid/details'")
		return nil
	}, []Interceptor{})
}
