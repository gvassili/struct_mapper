package struct_mapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type srcA struct {
	A int    `map:"path=B C""`
	B string `map:"path=B B"`
	C bool   `map:"path=B A"`
}

type srcB struct {
	A float64   `map:"path=A C"`
	B uint32    `map:"path=A B"`
	C time.Time `map:"path=A A"`
}

type DeepDecodeSrc struct {
	A srcA
	B srcB
}

type dstA struct {
	A time.Time
	B uint32
	C float64
}

type DstB struct {
	A bool
	B string
	C int
}

type DeepDecodeDst struct {
	A dstA
	B DstB
}

type DeepDecodePtrDst struct {
	A *dstA
	B *DstB
}

func TestStructMapDecoder_Decode(t *testing.T) {
	sm := New()
	ts, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00.000Z")
	src := DeepDecodeSrc{
		srcA{123, "text", true},
		srcB{1.23, 255, ts},
	}
	var dst DeepDecodeDst
	dec, err := sm.GetDecoder(&src, &dst)
	assert.NoError(t, err)
	err = dec.Decode(&src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, DeepDecodeDst{
		A: dstA{src.B.C, src.B.B, src.B.A},
		B: DstB{src.A.C, src.A.B, src.A.A},
	}, dst)
	var dstPtr DeepDecodePtrDst
	ptrDec, err := sm.GetDecoder(&src, &dstPtr)
	assert.NoError(t, err)
	err = ptrDec.Decode(&src, &dstPtr)
	assert.NoError(t, err)
	assert.Equal(t, DeepDecodePtrDst{
		A: &dstA{src.B.C, src.B.B, src.B.A},
		B: &DstB{src.A.C, src.A.B, src.A.A},
	}, dstPtr)
}
