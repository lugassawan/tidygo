# busywait

Forbids a `for` loop that spins on a plain variable: the body is empty, or is only `time.Sleep(...)` or `runtime.Gosched()`. Reading a variable that another goroutine writes, without synchronisation, is a data race, and the loop burns CPU (or polls on a timer) while it waits. Use a channel, `sync.Cond`, `sync.WaitGroup` or `context` instead.

The analyzer needs type information to recognise `time.Sleep` and `runtime.Gosched`. The loop must have a condition and no init or post statement. The condition must be built only from identifiers, selectors, parentheses, unary `!`/`-`/`+`, comparisons, `&&`/`||` and constants. Arithmetic (`x+1 < y`) and pointer dereferences (`*p`) are outside that set, so such loops are not reported. A condition that is a constant (`for true`) waits on no state and is not reported. Any function or method call, channel receive or index expression in the condition makes the loop exempt (type conversions and builtins do not count as calls, so `time.Sleep(time.Duration(n) * time.Millisecond)` is still a spin), which keeps polling idioms such as `atomic.Load*`, `ctx.Err()`, `time.Now().Before(d)` and `len(q)` out of scope. A bare `for {}` has no condition and is covered by staticcheck SA5002.

```go
// Bad
for !done {
}
for !s.ready {
	time.Sleep(10 * time.Millisecond)
}
for flag == false {
	runtime.Gosched()
}

// Good
<-doneCh
for atomic.LoadInt32(&state) == 0 { // an explicit atomic poll is not reported
	time.Sleep(time.Millisecond)
}
for i := 0; i < n; i++ { // has init and post
}
```

## Known limitations

A loop that does any other work in its body is never reported, even if it also spins on a plain variable. An atomic poll is not reported either, though a channel is usually still the better design.
