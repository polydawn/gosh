package iox

import (
	"bytes"
	"io"
	"strings"
)

/*
	Converts any of a range of data sources to an io.Reader interface, or
	an io.ReadCloser if appropriate.

	Readers will be produced from:
		string
		[]byte
		io.Reader
		bytes.Buffer
		<-chan string
		<-chan []byte
	ReadClosers will be produced from:
		chan string
		chan []byte

	An error of type ReaderUnrefinableFromInterface is thrown if an argument
	of any other type is given.
*/
func ReaderFromInterface(x interface{}) io.Reader {
	switch y := x.(type) {
	case string:
		return ReaderFromString(y)
	case []byte:
		return ReaderFromByteSlice(y)
	case io.Reader:
		return y
	case bytes.Buffer:
		return &y
	case <-chan string:
		return ReaderFromChanReadonlyString(y)
	case chan string:
		return ReaderFromChanString(y)
	case <-chan []byte:
		return ReaderFromChanReadonlyByteSlice(y)
	case chan []byte:
		return ReaderFromChanByteSlice(y)
	default:
		panic(ReaderUnrefinableFromInterface{wat: y})
	}
}

func ReaderFromString(str string) io.Reader {
	return strings.NewReader(str)
}

func ReaderFromByteSlice(bats []byte) io.Reader {
	return bytes.NewReader(bats)
}

func ReaderFromChanString(ch chan string) io.Reader {
	return &readerChanString{ch: ch}
}

type readerChanString struct {
	ch  chan string
	buf []byte
}

func (r *readerChanString) Read(p []byte) (n int, err error) {
	w := 0
	if len(r.buf) == 0 {
		// skip
	} else if len(p) >= len(r.buf) {
		// copy whole buffer out
		w = copy(p, r.buf)
		r.buf = r.buf[0:0]
	} else {
		// not room for the whole buffer; copy what there's room for, shift buf, return.
		w = copy(p, r.buf[:len(p)])
		r.buf = r.buf[len(p):0]
		return w, nil
	}

	str, open := <-r.ch
	bytes := []byte(str)
	w2 := copy(p, bytes)
	r.buf = bytes[w2:]

	if open || len(r.buf) > 0 {
		return w + w2, nil
	} else {
		return w + w2, io.EOF
	}
}

func (r *readerChanString) Close() error {
	close(r.ch)
	return nil
}

func ReaderFromChanReadonlyString(ch <-chan string) io.Reader {
	return &readerChanReadonlyString{ch: ch}
}

type readerChanReadonlyString struct {
	ch  <-chan string
	buf []byte
}

func (r *readerChanReadonlyString) Read(p []byte) (n int, err error) {
	w := 0
	if len(r.buf) == 0 {
		// skip
	} else if len(p) >= len(r.buf) {
		// copy whole buffer out
		w = copy(p, r.buf)
		r.buf = r.buf[0:0]
	} else {
		// not room for the whole buffer; copy what there's room for, shift buf, return.
		w = copy(p, r.buf[:len(p)])
		r.buf = r.buf[len(p):0]
		return w, nil
	}

	str, open := <-r.ch
	bytes := []byte(str)
	w2 := copy(p, bytes)
	r.buf = bytes[w2:]

	if open || len(r.buf) > 0 {
		return w + w2, nil
	} else {
		return w + w2, io.EOF
	}
}

func ReaderFromChanByteSlice(ch chan []byte) io.Reader {
	return &readerChanByteSlice{ch: ch}
}

type readerChanByteSlice struct {
	ch  chan []byte
	buf []byte
}

func (r *readerChanByteSlice) Read(p []byte) (n int, err error) {
	w := 0
	if len(r.buf) == 0 {
		// skip
	} else if len(p) >= len(r.buf) {
		// copy whole buffer out
		w = copy(p, r.buf)
		r.buf = r.buf[0:0]
	} else {
		// not room for the whole buffer; copy what there's room for, shift buf, return.
		w = copy(p, r.buf[:len(p)])
		r.buf = r.buf[len(p):0]
		return w, nil
	}

	bytes, open := <-r.ch
	w2 := copy(p, bytes)
	r.buf = bytes[w2:]

	if open || len(r.buf) > 0 {
		return w + w2, nil
	} else {
		return w + w2, io.EOF
	}
}

func (r *readerChanByteSlice) Close() error {
	close(r.ch)
	return nil
}

func ReaderFromChanReadonlyByteSlice(ch <-chan []byte) io.Reader {
	return &readerChanReadonlyByteSlice{ch: ch}
}

type readerChanReadonlyByteSlice struct {
	ch  <-chan []byte
	buf []byte
}

func (r *readerChanReadonlyByteSlice) Read(p []byte) (n int, err error) {
	w := 0
	if len(r.buf) == 0 {
		// skip
	} else if len(p) >= len(r.buf) {
		// copy whole buffer out
		w = copy(p, r.buf)
		r.buf = r.buf[0:0]
	} else {
		// not room for the whole buffer; copy what there's room for, shift buf, return.
		w = copy(p, r.buf[:len(p)])
		r.buf = r.buf[len(p):0]
		return w, nil
	}

	bytes, open := <-r.ch
	w2 := copy(p, bytes)
	r.buf = bytes[w2:]

	if open || len(r.buf) > 0 {
		return w + w2, nil
	} else {
		return w + w2, io.EOF
	}
}
