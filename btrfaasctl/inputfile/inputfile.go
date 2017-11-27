package inputfile

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Resolve resolves a file uri
func Resolve(uri string) ([]byte, error) {
	if _, err := os.Stat(uri); err == nil {
		return ioutil.ReadFile(uri)
	}
	if strings.HasPrefix(uri, "http") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return nil, errors.New("can not determine file " + uri)
}
