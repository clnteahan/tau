package zstd

/*
#cgo LDFLAGS: -lzstd
#include <zstd.h>
#include <stdlib.h>
*/

import "C"
import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

const DefaultCompressionLevel = 3
const (
	MinCompressionLevel = 1
	MaxCompressionLevel = 22
)

var errWriterClosed = errors.New("[ZSTD] Writer is closed")

type Writer struct {
	w       io.Writer
	cstream *C.ZSTD_CStream
	inBuf   []byte
	outBuf  []byte
	closed  bool
}

func NewWriter(w io.Writer) *Writer {
	nw, _ := NewWriterLevel(w, DefaultCompressionLevel)
	return nw
}

func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	if (level < MinCompressionLevel) || (level > MaxCompressionLevel) {
		return nil, errors.New(fmt.Sprintf("[ZSTD] Invalid compression level %d, must be in range [%d,%d]", level, MinCompressionLevel, MaxCompressionLevel))
	}

	cstream := *C.ZSTD_CStream

	if cstream == nil {
		return nil, fmt.Errorf("[ZSTD] Unable to create compression stream")
	}

	rc := C.ZSTD_initCStream(cstream, C.int(level))
	if C.ZSTD_isError(rc) != 0 {
		C.ZSTD_freeCStream(cstream)
		return nil, fmt.Errorf("[ZSTD] initCStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
	}

	outSize := int(C.ZSTD_CStreamOutSize())
	return &Writer{
		w:       w,
		cstream: cstream,
		inBuf:   make([]byte, 0, int(C.ZSTD_CStreamInSize())),
		outBuf:  make([]byte, outSize),
		closed:  false,
	}, nil

}

func (w *Writer) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errWriterClosed
	}
	if len(p) == 0 {
		return 0, nil
	}

	total := len(p)

	// Pin the Go slice to a stable address while the C code reads it.
	inPtr := unsafe.Pointer(&p[0])
	outPtr := unsafe.Pointer(&w.outBuf[0])
	outSize := C.size_t(len(w.outBuf))

	input := C.ZSTD_inBuffer{
		src:  inPtr,
		size: C.size_t(len(p)),
		pos:  0,
	}

	for input.pos < input.size {
		output := C.ZSTD_outBuffer{
			dst:  outPtr,
			size: outSize,
			pos:  0,
		}

		rc := C.ZSTD_compressStream(w.cstream, &output, &input)
		if C.ZSTD_isError(rc) != 0 {
			return 0, fmt.Errorf("[ZSTD] compressStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
		}

		if output.pos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(output.pos)]); err != nil {
				return 0, err
			}
		}
	}

	return total, nil
}

func (w *Writer) Flush() error {
	if w.closed {
		return errWriterClosed
	}
	return w.flush()
}

func (w *Writer) flush() error {
	outPtr := unsafe.Pointer(&w.outBuf[0])
	outSize := C.size_t(len(w.outBuf))

	for {
		output := C.ZSTD_outBuffer{
			dst:  outPtr,
			size: outSize,
			pos:  0,
		}

		remaining := C.ZSTD_flushStream(w.cstream, &output)
		if C.ZSTD_isError(remaining) != 0 {
			return fmt.Errorf("[ZSTD] flushStream: %s", C.GoString(C.ZSTD_getErrorName(remaining)))
		}

		if output.pos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(output.pos)]); err != nil {
				return err
			}
		}

		if remaining == 0 {
			break
		}
	}
	return nil
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	defer C.ZSTD_freeCStream(w.cstream)

	outPtr := unsafe.Pointer(&w.outBuf[0])
	outSize := C.size_t(len(w.outBuf))

	// endStream flushes and writes the frame epilogue.
	for {
		output := C.ZSTD_outBuffer{
			dst:  outPtr,
			size: outSize,
			pos:  0,
		}

		remaining := C.ZSTD_endStream(w.cstream, &output)
		if C.ZSTD_isError(remaining) != 0 {
			return fmt.Errorf("[ZSTD] endStream: %s", C.GoString(C.ZSTD_getErrorName(remaining)))
		}

		if output.pos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(output.pos)]); err != nil {
				return err
			}
		}

		if remaining == 0 {
			break
		}
	}
	return nil
}

func (w *Writer) Reset(dst io.Writer) error {
	if !w.closed {
		// Best-effort finalise the previous frame before discarding state.
		_ = w.Close()
	}

	w.w = dst
	w.closed = false

	rc := C.ZSTD_resetCStream(w.cstream, 0)
	if C.ZSTD_isError(rc) != 0 {
		return fmt.Errorf("zstd: resetCStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
	}
	return nil
}
