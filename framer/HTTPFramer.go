package framer

import (
	"bufio"
	"io"
	"net/http"
)

// HTTPFramer reads a HTTP response from the src and writes it to dest
type HTTPFramer struct{}

// Copy implements the Framer interface
func (framer *HTTPFramer) Copy(dest io.Writer, src io.Reader) error {
	resp, err := http.ReadResponse(bufio.NewReader(src), nil)
	if err != nil {
		return err
	}
	return resp.Write(dest)
}
