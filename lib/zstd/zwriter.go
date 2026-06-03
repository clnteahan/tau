package zstd

/*
#cgo LDFLAGS: -lzstd
#include <zstd.h>
#include <stdlib.h>

// Wrapper functions that construct the buffer structs entirely in C memory,
// so CGo never sees a Go struct containing Go pointers. This is the correct
// way to satisfy CGo's "no Go pointer in a value passed to C" rule.

static size_t _compressStream(ZSTD_CStream* cs,
                              void* dst, size_t dstCap, size_t* dstPos,
                              const void* src, size_t srcSize, size_t* srcPos) {
    ZSTD_outBuffer out = { dst, dstCap, *dstPos };
    ZSTD_inBuffer  in  = { src, srcSize, *srcPos };
    size_t rc = ZSTD_compressStream(cs, &out, &in);
    *dstPos = out.pos;
    *srcPos = in.pos;
    return rc;
}

static size_t _flushStream(ZSTD_CStream* cs,
                           void* dst, size_t dstCap, size_t* dstPos) {
    ZSTD_outBuffer out = { dst, dstCap, *dstPos };
    size_t rc = ZSTD_flushStream(cs, &out);
    *dstPos = out.pos;
    return rc;
}

static size_t _endStream(ZSTD_CStream* cs,
                         void* dst, size_t dstCap, size_t* dstPos) {
    ZSTD_outBuffer out = { dst, dstCap, *dstPos };
    size_t rc = ZSTD_endStream(cs, &out);
    *dstPos = out.pos;
    return rc;
}
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

var errWriterClosed = fmt.Errorf("[ZSTD] Writer is closed")

type Writer struct {
	w       io.Writer
	cstream *C.ZSTD_CStream
	outBuf  []byte
	closed  bool
}

func NewWriter(w io.Writer) *Writer {
	nw, _ := NewWriterLevel(w, DefaultCompressionLevel)
	return nw
}

func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	if level < MinCompressionLevel || level > MaxCompressionLevel {
		return nil, errors.New(fmt.Sprintf("[ZSTD] Invalid compression level %d, must be in range [%d,%d]", level, MinCompressionLevel, MaxCompressionLevel))
	}

	cstream := C.ZSTD_createCStream()
	if cstream == nil {
		return nil, fmt.Errorf("[ZSTD] Unable to create compression stream")
	}

	rc := C.ZSTD_initCStream(cstream, C.int(level))
	if C.ZSTD_isError(rc) != 0 {
		C.ZSTD_freeCStream(cstream)
		return nil, fmt.Errorf("[ZSTD] initCStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
	}

	return &Writer{
		w:       w,
		cstream: cstream,
		outBuf:  make([]byte, int(C.ZSTD_CStreamOutSize())),
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

	// C.CBytes copies p into C-heap memory, giving us a plain C pointer
	// with no Go pointer inside — fully satisfying CGo's rules.
	cIn := C.CBytes(p)
	defer C.free(cIn)

	outPtr := unsafe.Pointer(&w.outBuf[0])
	outCap := C.size_t(len(w.outBuf))
	srcSize := C.size_t(len(p))

	var srcPos, dstPos C.size_t

	for srcPos < srcSize {
		dstPos = 0
		rc := C._compressStream(w.cstream,
			outPtr, outCap, &dstPos,
			cIn, srcSize, &srcPos)
		if C.ZSTD_isError(rc) != 0 {
			return 0, fmt.Errorf("[ZSTD] compressStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
		}
		if dstPos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(dstPos)]); err != nil {
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
	outCap := C.size_t(len(w.outBuf))

	for {
		var dstPos C.size_t
		remaining := C._flushStream(w.cstream, outPtr, outCap, &dstPos)
		if C.ZSTD_isError(remaining) != 0 {
			return fmt.Errorf("[ZSTD] flushStream: %s", C.GoString(C.ZSTD_getErrorName(remaining)))
		}
		if dstPos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(dstPos)]); err != nil {
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
	outCap := C.size_t(len(w.outBuf))

	for {
		var dstPos C.size_t
		remaining := C._endStream(w.cstream, outPtr, outCap, &dstPos)
		if C.ZSTD_isError(remaining) != 0 {
			return fmt.Errorf("[ZSTD] endStream: %s", C.GoString(C.ZSTD_getErrorName(remaining)))
		}
		if dstPos > 0 {
			if _, err := w.w.Write(w.outBuf[:int(dstPos)]); err != nil {
				return err
			}
		}
		if remaining == 0 {
			break
		}
	}
	return nil
}

// Reset re-targets the Writer at a new destination without allocating a new
// cstream. Uses the modern ZSTD_CCtx_reset API (zstd >= 1.4.0).
func (w *Writer) Reset(dst io.Writer) error {
	if !w.closed {
		_ = w.Close()
	}
	w.w = dst
	w.closed = false

	rc := C.ZSTD_CCtx_reset(w.cstream, C.ZSTD_reset_session_only)
	if C.ZSTD_isError(rc) != 0 {
		return fmt.Errorf("[ZSTD] CCtx_reset: %s", C.GoString(C.ZSTD_getErrorName(rc)))
	}
	return nil
}
