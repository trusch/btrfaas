package forwarder

import (
	"context"
	"io"
	"net/http"
)

// HTTPRunnable is a Runnable which does an HTTP request for its work (to be used with openfaas)
type HTTPRunnable struct {
	url string
}

// NewHTTPRunnable returns a new http runnable for a given host port combination
func NewHTTPRunnable(url string) *HTTPRunnable {
	return &HTTPRunnable{url}
}

// Run implements the runnable interface
func (r *HTTPRunnable) Run(ctx context.Context, options map[string]string, input io.Reader, output io.Writer) error {
	req, err := http.NewRequest("POST", r.constructURL(options), input)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(output, resp.Body)
	return err
}

func (r *HTTPRunnable) constructURL(options map[string]string) string {
	res := r.url
	if options != nil && len(options) > 0 {
		res += "?"
	}
	for k, v := range options {
		res += k + "=" + v + "&"
	}
	if options != nil && len(options) > 0 {
		res = res[:len(res)-1] // remove last ampersant
	}
	return res
}
