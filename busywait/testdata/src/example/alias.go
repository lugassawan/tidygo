package example

import tm "time"

func AliasedSleep(d tm.Duration) {
	for done == false { // want `loop busy-waits on done == false; use a channel, sync.Cond or context instead`
		tm.Sleep(d)
	}
}
