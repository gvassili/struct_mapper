package struct_mapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SimpleSrc struct {
	A string `map:"C"`
	B string `map:"B"`
	C string `map:"A"`
}

type SimpleDst struct {
	A string
	B string
	C string
}

type FlatSrc struct {
	AA string `map:"path=B C"`
	AB string `map:"path=B B"`
	AC string `map:"path=B A"`
	BA string `map:"path=A C"`
	BB string `map:"path=A B"`
	BC string `map:"path=A A"`
}

type DeepSrc struct {
	A struct {
		A string `map:"path=B C"`
		B string `map:"path=B B"`
		C string `map:"path=B A"`
	}
	B struct {
		A string `map:"path=A C"`
		B string `map:"path=A B"`
		C string `map:"path=A A"`
	}
}

type DeepDst struct {
	A struct {
		A string
		B string
		C string
	}
	B struct {
		A string
		B string
		C string
	}
}

type DeepPtrDst struct {
	A *struct {
		A string
		B string
		C string
	}
	B *struct {
		A string
		B string
		C string
	}
}

func TestStructMapper_GetDecoder(t *testing.T) {
	sm := New()
	_, err := sm.GetDecoder(&SimpleSrc{}, &SimpleDst{})
	assert.NoError(t, err)
	_, err = sm.GetDecoder(&FlatSrc{}, &DeepDst{})
	assert.NoError(t, err)
	_, err = sm.GetDecoder(&DeepSrc{}, &DeepDst{})
	assert.NoError(t, err)
	_, err = sm.GetDecoder(&DeepSrc{}, &DeepPtrDst{})
	assert.NoError(t, err)
}
