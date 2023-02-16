package plane_test

import (
	"testing"

	plane "./"
)

const (
	defaultRadius  = 1000
	defaultWorkers = 20
)

type signed interface {
	int | int8 | int16 | int32 | int64
}

type unsigned interface {
	uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

type integer interface {
	signed | unsigned
}

func newPlane() *plane.Plane {
	return plane.InitNewPlane(defaultRadius, defaultWorkers)
}

func isEmpty[T integer](a []T) bool {
	if a == nil {
		return true
	}

	var allZeroValues bool = true
	for _, n := range a {
		if n != 0 {
			allZeroValues = false
			break
		}
	}

	return allZeroValues || len(a) == 0
}

func TestStart(t *testing.T) {
	pln := newPlane()
	pln.Start()
	if !pln.IsRunning() {
		t.Error("calculations are expected to be running")
	}
	pln.Stop()
}

func TestStop(t *testing.T) {
	pln := newPlane()
	pln.Start()
	pln.Stop()
	if pln.IsRunning() {
		t.Error("calculations are expected to be stopped")
	}
}

func TestIsFinished(t *testing.T) {
	pln := newPlane()
	if pln.IsFinished() {
		t.Error("calculations are expected not to be finished")
	}
	pln.Start()
	<-pln.NotifyOnFinish()
	if !pln.IsFinished() {
		t.Error("calculations are expected to be finished")
	}
}

func TestNotify(t *testing.T) {
	pln := newPlane()
	pln.Start()
	<-pln.NotifyOnFinish()
	if pln.IsRunning() {
		t.Error("calculations are expected to be stopped")
	}
	if !pln.IsFinished() {
		t.Error("calculations are expected to be finished")
	}
}

func TestResultsWithoutRunning(t *testing.T) {
	pln := newPlane()
	res := pln.GetResults()
	if !isEmpty(res) {
		t.Errorf("slice is not empty: %v\n", res)
	}
}

func TestResultsWhileRunning(t *testing.T) {
	pln := newPlane()
	pln.Start()
	res := pln.GetResults()
	if res != nil {
		t.Errorf("slice is expected to be nil: %v\n", res)
	}
	pln.Stop()
}
