package iox

import (
	"bytes"
	"io"
)

/*
	Converts any of a range of data sinks to an io.Writer interface, or
	an io.WriteCloser if appropriate.

	Writers will be produced from:
		io.Writer
		bytes.Buffer
	WriteClosers will be produced from:
		<-chan string
		chan string
		<-chan []byte
		chan []byte

	An error of type WriterUnrefinableFromInterface is thrown if an argument
	of any other type is given.
*/
func WriterFromInterface(x interface{}) io.Writer {
	switch y := x.(type) {
	case io.Writer:
		return y
	case bytes.Buffer:
		return &y
	case chan<- string:
		return WriterToChanString(y)
	case chan string:
		return WriterToChanString(y)
	case chan<- []byte:
		return WriterToChanByteSlice(y)
	case chan []byte:
		return WriterToChanByteSlice(y)
	default:
		panic(WriterUnrefinableFromInterface{wat: y})
	}
}

func WriterToChanString(ch chan<- string) io.Writer {
	return &writerChanString{ch: ch}
}

type writerChanString struct {
	ch chan<- string
}

func (r *writerChanString) Write(p []byte) (n int, err error) {
	defer func() {
		if e := recover(); e != nil {
			n = 0
			err = io.EOF
		}
	}()

	r.ch <- string(p)
	return len(p), nil
}

func (r *writerChanString) Close() error {
	close(r.ch)
	return nil
}

func WriterToChanByteSlice(ch chan<- []byte) io.Writer {
	return &writerChanByteSlice{ch: ch}
}

type writerChanByteSlice struct {
	ch chan<- []byte
}

func (r *writerChanByteSlice) Write(p []byte) (n int, err error) {
	defer func() {
		if e := recover(); e != nil {
			n = 0
			err = io.EOF
		}
	}()

	r.ch <- p
	return len(p), nil
}

func (r *writerChanByteSlice) Close() error {
	close(r.ch)
	return nil
}
