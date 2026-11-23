package example

import (
	"runtime"
	"time"
)

func EmptySpin() {
	for !done { // want `loop busy-waits on !done; use a channel, sync.Cond or context instead`
	}
}

func SleepPoll() {
	for !done { // want `loop busy-waits on !done; use a channel, sync.Cond or context instead`
		time.Sleep(10 * time.Millisecond)
	}
}

func GoschedPoll(s *state) {
	for s.ready == false { // want `loop busy-waits on s.ready == false; use a channel, sync.Cond or context instead`
		runtime.Gosched()
	}
}

func CompoundCondition() {
	for x < 3 && !flag { // want `loop busy-waits on x < 3 && !flag; use a channel, sync.Cond or context instead`
	}
}

func ParenthesisedCondition() {
	for !(done || flag) { // want `loop busy-waits on !\(done \|\| flag\); use a channel, sync.Cond or context instead`
	}
}

func NegatedOperand() {
	for -x > 0 { // want `loop busy-waits on -x > 0; use a channel, sync.Cond or context instead`
	}
}

func InsideClosure() {
	go func() {
		for !done { // want `loop busy-waits on !done; use a channel, sync.Cond or context instead`
		}
	}()
}
