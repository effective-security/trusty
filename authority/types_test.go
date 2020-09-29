package authority

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration_String(t *testing.T) {
	f := func(d time.Duration, exp string) {
		actual := Duration(d).String()
		require.Equal(t, exp, actual)
	}
	f(time.Second, "1s")
	f(time.Second*30, "30s")
	f(time.Minute, "1m0s")
	f(time.Second*90, "1m30s")
	f(0, "0s")
}

func TestDuration_JSON(t *testing.T) {
	f := func(d time.Duration, exp string) {
		v := Duration(d)
		bytes, err := json.Marshal(&v)
		require.NoError(t, err)
		require.Equal(t, exp, string(bytes))
		var decoded Duration

		err = json.Unmarshal(bytes, &decoded)
		require.NoError(t, err)

		assert.Equal(t, v, decoded)
	}
	f(0, `"0s"`)
	f(time.Second, `"1s"`)
	f(time.Minute*5, `"5m0s"`)
	f(time.Second*90, `"1m30s"`)
	f(time.Hour*2, `"2h0m0s"`)
	f(time.Millisecond*10, `"10ms"`)
}

func TestDuration_JSONDecode(t *testing.T) {
	f := func(j string, exp time.Duration) {
		var act Duration
		err := json.Unmarshal([]byte(j), &act)
		require.NoError(t, err)
		assert.Equal(t, exp, act.TimeDuration())
	}
	f(`"5m"`, time.Minute*5)
	f(`120`, time.Second*120)
	f(`0`, 0)
	f(`"1m5s"`, time.Second*65)
}
