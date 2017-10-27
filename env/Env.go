package env

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Env represents a key/value mapping of environment variables
type Env map[string]string

// ReadOSEnvironment parses the operating systems environment variables
func (env Env) ReadOSEnvironment() error {
	e := os.Environ()
	for _, pair := range e {
		parts := strings.Split(pair, "=")
		if len(parts) < 2 {
			return errors.New("malformed environemt entry: " + pair)
		}
		env[parts[0]] = strings.Join(parts[1:], "=")
	}
	return nil
}

// AddFromHTTPRequest adds environment variables from an HTTP request
func (env Env) AddFromHTTPRequest(r *http.Request) {
	for k, v := range r.Header {
		kv := fmt.Sprintf("Http_%s=%s", strings.Replace(k, "-", "_", -1), v[0])
		parts := strings.Split(kv, "=")
		env[parts[0]] = strings.Join(parts[1:], "=")
	}
	env["Http_Method"] = r.Method
	if len(r.URL.RawQuery) > 0 {
		env["Http_Query"] = r.URL.RawQuery
	}
	if len(r.URL.Path) > 0 {
		env["Http_Path"] = r.URL.Path
	}
}

// Copy returns a copy of the current environment
func (env Env) Copy() Env {
	res := make(Env)
	for k, v := range env {
		res[k] = v
	}
	return res
}

// ToSlice returns a env slice usable in exec.Cmd
func (env Env) ToSlice() []string {
	res := make([]string, len(env))
	i := 0
	for k, v := range env {
		res[i] = k + "=" + v
		i++
	}
	return res
}
