### Build Status
![Build Status](https://travis-ci.org/asvins/router.svg?branch=master)

# router
### A simple golang HTTP router

## Usage

### Simple Route
Simple handler registration for the route /api/product
```go
	import("github.com/asvins/router/logger")
	...
	r := router.NewRouter()
	r.AddRoute("/api/product", router.GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/api/users'")
	})
	...
```

### Simple Route that lets the router write the error on response
```go
	import(
		routerErrors "github.com/asvins/router/errors"
	)	
	r.Handle("/handler/unauthorized", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		return routerErrors.Unauthorized("You shall not pass")
	}, []Interceptor{})
```
#### IMPORTANT: DO NOT WRITE INTO THE RESPONSE WRITER IF USE USE r.Handle()

### Simple route with <:param> notation
```go
	import(
		routerErrors "github.com/asvins/router/errors"
	)	
	r.Handle("/user/:uid/details/:did", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		params := rq.URL.Query()
		uid := params.Get("uid") // will return uid as stirng
		did := params.Get("did") // will return did as string

		// be awsome

		return nil
	}, []Interceptor{})

```

### Simple route with <:param> notation and query string
```go
	import(
		routerErrors "github.com/asvins/router/errors"
	)	
	
	r.Handle("/user/:uid/appointment?aid=1234", GET, func(w http.ResponseWriter, rq *http.Request) routerErrors.Http {
		params := rq.URL.Query()
		uid := params.Get("uid") // will return uid as stirng
		aid := params.Get("aid") // will return did as string

		// be awsome

		return nil
	}, []Interceptor{})

```


### Route with specific Interceptor
The route /api/user will be intercepter by the logger interceptor
```go
	...
	r.AddRoute("/api/user", router.GET, func(w http.ResponseWriter, rq *http.Request) {
		fmt.Fprint(w, "Request made to '/'")
	}, logger.NewLogger())
	...
```

### Base Interceptor
All requests that hit /api/... will be intercepted by the logger interceptor
```go
	...
	r.AddBaseInterceptor("/api", logger.NewLogger)
	...
```

for the specific and base interceptor registration examples given, the logger interceptor is defined as:
```go
package logger

import (
	"fmt"
	"net/http"
)

//Logger is a dummy example of possible interceptor
type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

//Intercept is the Interceptor interface implementation
func (l Logger) Intercept(rw http.ResponseWriter, r *http.Request) error {
	fmt.Println("[LOG] Request from: ", r.RemoteAddr, " ", r.Method, " ", r.URL.String())
	return nil
}
```

