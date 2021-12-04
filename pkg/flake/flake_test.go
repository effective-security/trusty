package flake

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sf *Flake

var startTime int64
var machineID uint64

func init() {
	var st Settings
	st.StartTime = time.Now()

	sf = NewIDGenerator(st).(*Flake)
	if sf == nil {
		panic("sonyflake not created")
	}

	startTime = toFlakeTime(st.StartTime)

	ip, _ := defaultMachineID()
	machineID = uint64(ip)
}

func TestFlakeOnce(t *testing.T) {
	sleepTime := uint64(50)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)

	id := sf.NextID()
	parts := Decompose(id)
	t.Logf("parts: %+v", parts)

	actualMSB := parts["msb"]
	assert.Equal(t, uint64(0), actualMSB)

	actualTime := parts["time"]
	if actualTime < sleepTime || actualTime > sleepTime+3 {
		t.Errorf("unexpected time: %d", actualTime)
	}

	actualSequence := parts["sequence"]
	assert.Equal(t, uint64(0), actualSequence)

	actualMachineID := parts["machine-id"]
	assert.Equal(t, uint64(machineID), uint64(actualMachineID))
}

func currentTime() int64 {
	return toFlakeTime(time.Now())
}

func TestFlakeFor10Sec(t *testing.T) {
	var numID uint32
	var lastID uint64
	var maxSequence uint64

	initial := currentTime()
	current := initial
	for current-initial < 10000 {
		id := sf.NextID()
		parts := Decompose(id)
		numID++

		if id <= lastID {
			t.Fatal("duplicated id")
		}
		lastID = id

		current = currentTime()

		actualMSB := parts["msb"]
		require.Equal(t, uint64(0), actualMSB)

		actualTime := int64(parts["time"])
		overtime := startTime + actualTime - current
		if overtime > 0 {
			t.Errorf("unexpected overtime: %d", overtime)
		}

		actualSequence := parts["sequence"]
		if maxSequence < actualSequence {
			maxSequence = actualSequence
		}

		actualMachineID := parts["machine-id"]
		require.Equal(t, uint64(machineID), uint64(actualMachineID))
	}
	/*
		if maxSequence < 90 {
			t.Errorf("unexpected max sequence: %d", maxSequence)
		}
	*/
	t.Logf("max sequence: %d", maxSequence)
	t.Logf("number of id: %d", numID)
}

func TestFlakeInParallel(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	t.Logf("number of cpu: %d", numCPU)

	consumer := make(chan uint64)

	const numID = 10000
	generate := func() {
		for i := 0; i < numID; i++ {
			consumer <- sf.NextID()
		}
	}

	const numGenerator = 10
	for i := 0; i < numGenerator; i++ {
		go generate()
	}

	var maxSequence uint64

	set := mapset.NewSet()
	for i := 0; i < numID*numGenerator; i++ {
		id := <-consumer
		if set.Contains(id) {
			t.Fatal("duplicated id")
		} else {
			set.Add(id)
		}

		parts := Decompose(id)

		actualSequence := parts["sequence"]
		if maxSequence < actualSequence {
			maxSequence = actualSequence
		}
	}
	t.Logf("number of id: %d", set.Cardinality())
	t.Logf("max sequence: %d", maxSequence)
}

func TestNilFlake(t *testing.T) {
	var startInFuture Settings
	startInFuture.StartTime = time.Now().Add(time.Duration(1) * time.Minute)
	assert.Panics(t, func() {
		NewIDGenerator(startInFuture)
	})

	var noMachineID Settings
	noMachineID.MachineID = func() (uint8, error) {
		return 0, fmt.Errorf("no machine id")
	}

	assert.Panics(t, func() {
		NewIDGenerator(noMachineID)
	})

	var invalidMachineID Settings
	invalidMachineID.CheckMachineID = func(uint8) bool {
		return false
	}
	assert.Panics(t, func() {
		NewIDGenerator(invalidMachineID)
	})
}

func pseudoSleep(period time.Duration) {
	sf.startTime -= int64(period) / flakeTimeUnit
}

func TestNextIDError(t *testing.T) {
	year := time.Duration(365*24) * time.Hour

	for i := 1; i < 35; i++ {
		t.Logf("over %d year", i)
		pseudoSleep(year)
		sf.NextID()
	}

	pseudoSleep(time.Duration(1) * year)
	assert.Panics(t, func() {
		sf.NextID()
	})
}
