package framer

import "io"

// Framer is an interface to specify different framing formats when copy data from one reader to a writer
type Framer interface {
	Copy(dest io.Writer, src io.Reader) error
}
