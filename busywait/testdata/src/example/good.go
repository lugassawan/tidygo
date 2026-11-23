package example

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"
)

func backoff() time.Duration { return time.Millisecond }

// The empty-condition spin belongs to staticcheck SA5002.
func BareSpin() {
	for {
	}
}

func AtomicPoll() {
	for atomic.LoadInt32(&counter) == 0 {
		time.Sleep(time.Millisecond)
	}
}

func ContextPoll(ctx context.Context) {
	for ctx.Err() == nil {
		time.Sleep(time.Millisecond)
	}
}

func DeadlinePoll(deadline time.Time) {
	for time.Now().Before(deadline) {
	}
}

func ReceiveInCondition() {
	for !<-ready {
	}
}

func IndexInCondition() {
	for !flags[0] {
	}
}

func LengthPoll() {
	for len(queue) == 0 {
	}
}

func WithInit(n int) {
	for i := 0; i < n; {
	}
}

func WithPost(n int) {
	i := 0
	for ; i < n; i++ {
	}
}

func Counting(n int) {
	for i := 0; i < n; i++ {
	}
}

func TwoStatements() {
	for !done {
		time.Sleep(time.Millisecond)
		counter++
	}
}

func SleepArgumentCallsFunc() {
	for !done {
		time.Sleep(backoff())
	}
}

func RealWork() {
	for !done {
		counter++
	}
}

func OtherRuntimeCall() {
	for !done {
		runtime.GC()
	}
}

func ReceiveStatement() {
	for !done {
		<-ready
	}
}

func LocalSleepMethod(f fake) {
	for !done {
		f.Sleep()
	}
}

func LocalGoschedMethod(f fake) {
	for !done {
		f.Gosched()
	}
}

func RangeLoop(ch chan int) {
	for range ch {
	}
}
