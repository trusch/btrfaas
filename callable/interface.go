package callable

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/trusch/frunner/env"
)

// Callable is the interface of all "functions" which can be run by the frunner
type Callable interface {
	Call(input io.Reader, env env.Env, output io.Writer) chan *CallError
	Stop() error
	Copy() Callable
}

// CallError contains the error returned while executing the callable and a reader to the additional error output
type CallError struct {
	Err    error
	Stderr io.Reader
}

func (e *CallError) Error() string {
	bs := []byte{}
	if e.Stderr != nil {
		bs, _ = ioutil.ReadAll(e.Stderr)
	}
	return fmt.Sprintf("%v: %v", e.Err, string(bs))
}
