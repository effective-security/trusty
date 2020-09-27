package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_ParseBuild(t *testing.T) {
	v := Info{Build: "1.2.114"}
	v.PopulateFromBuild()
	assert.Equal(t, uint(1), v.Major, "Major version")
	assert.Equal(t, uint(2), v.Minor, "MajMinoror version")
	assert.Equal(t, uint(114), v.Commit, "Commit")
	assert.Equal(t, float32(1.2*1000000+114), v.Float(), "Float")

	v = Info{Build: "1.2.114-dirty"}
	v.PopulateFromBuild()
	assert.Equal(t, uint(1), v.Major, "Major version")
	assert.Equal(t, uint(2), v.Minor, "MajMinoror version")
	assert.Equal(t, uint(114), v.Commit, "Commit")
	assert.Equal(t, float32(1.2*1000000+114), v.Float(), "Float")
}

func TestInfo_GreaterOrEqual(t *testing.T) {
	v01 := Info{0, 1, 3, "", "go1.5", float32(0.1*1000000 + 3)}
	v02 := Info{0, 2, 3, "", "go1.5", float32(0.2*1000000 + 3)}
	v10 := Info{1, 0, 3, "", "go1.5", float32(1.0*1000000 + 3)}
	v12 := Info{1, 2, 3, "", "go1.5", float32(1.2*1000000 + 3)}
	v20 := Info{2, 0, 3, "", "go1.5", float32(2.0*1000000 + 3)}
	f := func(v, other Info, expected bool) {
		act := v.GreaterOrEqual(other)
		if act != expected {
			t.Errorf("%v GreaterOrEqual (%v) return wrong result of %v, expecting %v", v, other, act, expected)
		}
	}
	f(v01, v01, true)
	f(v02, v01, true)
	f(v10, v01, true)
	f(v20, v12, true)
	f(v02, v10, false)
	f(v01, v02, false)
}
