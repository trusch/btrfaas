package framer

import (
	"encoding/json"
	"io"
)

// JSONFramer reads JSON from the src and writes it to dest
type JSONFramer struct{}

// Copy implements the Framer interface
func (framer *JSONFramer) Copy(dest io.Writer, src io.Reader) error {
	decoder := json.NewDecoder(src)
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return err
	}
	encoder := json.NewEncoder(dest)
	return encoder.Encode(data)
}
