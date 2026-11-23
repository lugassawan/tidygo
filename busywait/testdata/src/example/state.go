package example

// Shared state for the busy-wait fixtures. None of it is synchronised, which
// is exactly what the analyzer is meant to flag when a loop spins on it.
var (
	done    bool
	flag    bool
	x       int
	counter int32
	queue   []int
	flags   []bool
	ready   chan bool
)

type state struct{ ready bool }

type fake struct{}

func (fake) Sleep()   {}
func (fake) Gosched() {}
