package logger

import (
	"fmt"
	"net/http"

	"github.com/asvins/router/errors"
)

//Logger is a dummy example of possible interceptor
type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

//Intercept is the Interceptor interface implementation
func (l Logger) Intercept(rw http.ResponseWriter, r *http.Request) errors.Http {
	fmt.Println("[LOG] Request from: ", r.RemoteAddr, " ", r.Method, " ", r.URL.String())
	return nil
}
