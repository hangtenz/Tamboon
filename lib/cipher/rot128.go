package cipher

import (
	"bufio"
	"io"
)

// Rot128Reader implements io.Reader that transforms
type Rot128Reader struct{ reader bufio.Reader }

func NewRot128Reader(r bufio.Reader) (*Rot128Reader, error) {
	return &Rot128Reader{reader: r}, nil
}

func (r *Rot128Reader) Read(p []byte) (int, error) {
	if n, err := r.reader.Read(p); err != nil {
		return n, err
	} else {
		rot128(p[:n])
		return n, nil
	}
}

func (r *Rot128Reader) ReadBytes(delim byte) (p []byte, err error) {
	p, err = r.reader.ReadBytes(delim)
	if err != nil {
		return nil, err
	}
	rot128(p)
	return p, nil
}

type Rot128Writer struct {
	writer io.Writer
	buffer []byte // not thread-safe
}

func NewRot128Writer(w io.Writer) (*Rot128Writer, error) {
	return &Rot128Writer{
		writer: w,
		buffer: make([]byte, 4096, 4096),
	}, nil
}

func (w *Rot128Writer) Write(p []byte) (int, error) {
	n := copy(w.buffer, p)
	rot128(w.buffer[:n])
	return w.writer.Write(w.buffer[:n])
}

func rot128(buf []byte) {
	for idx := range buf {
		buf[idx] += 128
	}
}
