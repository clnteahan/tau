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

// Wrapper that keeps all buffer structs on the C stack, so no Go struct
// containing a Go pointer is ever passed across the CGo boundary.
static size_t _decompressStream(ZSTD_DStream* ds,
                                void*       dst, size_t dstCap, size_t* dstPos,
                                const void* src, size_t srcSize, size_t* srcPos) {
    ZSTD_outBuffer out = { dst, dstCap, *dstPos };
    ZSTD_inBuffer  in  = { src, srcSize, *srcPos };
    size_t rc = ZSTD_decompressStream(ds, &out, &in);
    *dstPos = out.pos;
    *srcPos = in.pos;
    return rc;
}
*/
import "C"

var errReaderClosed = errors.New("[ZSTD] Reader is closed")

type Reader struct {
	r       io.Reader
	dstream *C.ZSTD_DStream
	inBuf   []byte // compressed data read from r
	outBuf  []byte // decompressed data waiting to be copied to caller
	outPos  int    // read cursor into outBuf
	outFill int    // how many bytes in outBuf are valid
	inPos   int    // how many bytes in inBuf have been consumed by zstd
	inFill  int    // how many bytes in inBuf were read from r
	closed  bool
}

func NewReader(r io.Reader) (*Reader, error) {
	// FIX 1: nil-check before any use.
	dstream := C.ZSTD_createDStream()
	if dstream == nil {
		return nil, fmt.Errorf("[ZSTD] Unable to create decompression stream")
	}

	rc := C.ZSTD_initDStream(dstream)
	if C.ZSTD_isError(rc) != 0 {
		C.ZSTD_freeDStream(dstream)
		return nil, fmt.Errorf("[ZSTD] initDStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
	}

	return &Reader{
		r:       r,
		dstream: dstream,
		inBuf:   make([]byte, int(C.ZSTD_DStreamInSize())),
		outBuf:  make([]byte, int(C.ZSTD_DStreamOutSize())),
	}, nil
}

func (r *Reader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, errReaderClosed
	}
	if len(p) == 0 {
		return 0, nil
	}

	copied := 0

	for copied < len(p) {
		// FIX 3: Drain any previously decompressed bytes first.
		if r.outPos < r.outFill {
			n := copy(p[copied:], r.outBuf[r.outPos:r.outFill])
			r.outPos += n
			copied += n
			continue
		}
		// outBuf is empty — reset cursors.
		r.outPos = 0
		r.outFill = 0

		// FIX 7: Only read new compressed data when inBuf is fully consumed.
		if r.inPos >= r.inFill {
			n, err := r.r.Read(r.inBuf)
			if err != nil && err != io.EOF {
				return copied, err
			}
			if n == 0 {
				// Underlying reader is done; surface EOF to caller.
				if copied > 0 {
					return copied, nil
				}
				return 0, io.EOF
			}
			r.inPos = 0
			r.inFill = n
		}

		// FIX 2: Use C wrapper — no Go struct with Go pointer crosses the boundary.
		// FIX 4: Use output.pos (actual bytes written) not buffer capacity.
		// FIX 5: Decompress into outBuf, not directly into p.
		cIn := C.CBytes(r.inBuf[r.inPos:r.inFill])
		inSize := C.size_t(r.inFill - r.inPos)
		outPtr := unsafe.Pointer(&r.outBuf[0])
		outCap := C.size_t(len(r.outBuf))

		var srcPos, dstPos C.size_t
		rc := C._decompressStream(r.dstream,
			outPtr, outCap, &dstPos,
			cIn, inSize, &srcPos)
		C.free(cIn)

		if C.ZSTD_isError(rc) != 0 {
			return copied, fmt.Errorf("[ZSTD] decompressStream: %s", C.GoString(C.ZSTD_getErrorName(rc)))
		}

		r.inPos += int(srcPos)
		r.outFill = int(dstPos)
	}

	return copied, nil
}

func (r *Reader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	C.ZSTD_freeDStream(r.dstream)
	return nil
}
