package framer

import (
	"bufio"
	"io"
	"log"
)

// LineFramer reads a line from the src and writes it to dest
type LineFramer struct{}

// Copy implements the Framer interface
func (framer *LineFramer) Copy(dest io.Writer, src io.Reader) error {
	buf := bufio.NewReader(src)
	log.Print("start reading line...")
	line, err := buf.ReadBytes('\n')
	if err != nil {
		return err
	}
	_, err = dest.Write(line)
	return err
}
