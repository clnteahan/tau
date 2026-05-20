package zstd

import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

/*
#cgo LDFLAGS: -lzstd
#include <zstd.h>
#include <stdlib.h>
*/

import "C"

var errReaderClosed = errors.New("[ZSTD] Reader is closed")

type Reader struct {
	r       io.Reader
	dstream *C.zstd_stream
	inBuf   []byte
	outBuf  []byte
	closed  bool
}

func NewReader(r io.Reader) (*Reader, error) {
	dstream := *C.ZSTD_CStream

	if dstream == nil {
		return nil, fmt.Errorf("[ZSTD] Unable to create decompression stream")
	}

	rd := C.initDStream(dstream)

	if C.ZSTD_isError(rd) != 0 {
		C.ZSTD_freeDStream(dstream)
		return nil, fmt.Errorf("[ZSTD] initDStream: %s", C.GoString(C.ZSTD_getErrorName(rd)))
	}

	inSize := C.ZSTD_DStreamInSize()
	outSize := C.ZSTD_DStreamOutSize()

	return &Reader{
		r:       r,
		dstream: dstream,
		inBuf:   make([]byte, inSize),
		outBuf:  make([]byte, outSize),
		closed:  false,
	}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, errWriterClosed
	}
	length := len(p)
	if length == 0 {
		return 0, nil
	}

	progress := 0

	for progress < length {
		inSize, err := r.Read(r.inBuf)
		if err != nil && err != io.EOF {
			return inSize, err
		}
		inPtr := unsafe.Pointer(&r.inBuf[0])
		outPtr := unsafe.Pointer(&p[0])
		outSize := C.size_t(len(r.outBuf))

		input := C.ZSTD_inBuffer{
			src:  inPtr,
			size: C.size_t(inSize),
			pos:  0,
		}
		output := C.ZSTD_outBuffer{
			dst:  outPtr,
			size: outSize,
			pos:  0,
		}
		rd := C.ZSTD_compressStream(r.dstream, &output, &input)

		if C.ZSTD_isError(rd) != 0 {
			return 0, fmt.Errorf("[ZSTD] decompressStream: %s", C.GoString(C.ZSTD_getErrorName(rd)))
		}

		if output.pos > 0 {
			if _, err := r.r.Write(w.outBuf[:int(output.pos)]); err != nil {
				return 0, err
			}
		}
	}

	return r.r.Read(p)
}

func (r *Reader) Close() error {
	if r.closed {
		return errReaderClosed
	}
	r.closed = true
	C.ZSTD_freeDStream(r.dstream)
	return nil
}
