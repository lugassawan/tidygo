package example

import (
	"context"
	stdctx "context"
	"time"
)

type holder struct {
	cancel context.CancelFunc
}

var stored context.CancelFunc

func use(context.Context) {}

func keep(func()) {}

func done() {}

// ---- reported: the cancel func is not deferred on the next statement ----

func lostWithCancel() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func lostWithTimeout() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	) // want `cancel func from context.WithTimeout should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func lostWithDeadline() {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now(),
	) // want `cancel func from context.WithDeadline should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func lostWithCancelCause() {
	ctx, cancel := context.WithCancelCause(
		context.Background(),
	) // want `cancel func from context.WithCancelCause should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel(nil)
}

func lostWithTimeoutCause() {
	ctx, cancel := context.WithTimeoutCause(
		context.Background(),
		time.Second,
		nil,
	) // want `cancel func from context.WithTimeoutCause should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func lostWithDeadlineCause() {
	ctx, cancel := context.WithDeadlineCause(
		context.Background(),
		time.Now(),
		nil,
	) // want `cancel func from context.WithDeadlineCause should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func deferNotImmediate() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	defer cancel()
}

func unrelatedDeferNext() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	defer done()
	use(ctx)
	cancel()
}

func assignForm() {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func blankContext() {
	_, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	cancel()
}

func customName() {
	ctx, stop := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer stop\(\)"`
	use(ctx)
	stop()
}

func aliasImport() {
	ctx, cancel := stdctx.WithTimeout(
		context.Background(),
		time.Second,
	) // want `cancel func from context.WithTimeout should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
}

func inCaseClause(n int) {
	switch n {
	case 1:
		ctx, cancel := context.WithCancel(
			context.Background(),
		) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
		use(ctx)
		cancel()
	case 2:
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		use(ctx)
	}
}

func nestedLiteral() {
	lost := func() {
		ctx, cancel := context.WithCancel(
			context.Background(),
		) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
		use(ctx)
		cancel()
	}
	lost()

	fine := func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		use(ctx)
	}
	fine()
}

func outerLostInnerFine() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	func() {
		inner, cancelInner := context.WithCancel(ctx)
		defer cancelInner()
		use(inner)
	}()
	cancel()
}

// A deferred closure whose only call targets a parameter named cancel does not
// cancel the context function's cancel: identity, not the name, decides.
func shadowedInDefer() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	defer func(cancel func()) { cancel() }(func() {})
	use(ctx)
	cancel()
}

// Passing a different variable that happens to share the name is not an escape.
func shadowedElsewhere() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	) // want `cancel func from context.WithCancel should be deferred immediately: add "defer cancel\(\)"`
	use(ctx)
	cancel()
	{
		cancel := func() {}
		keep(cancel)
	}
}

// ---- accepted: the cancel func is deferred on the next statement ----

func deferredWithCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	use(ctx)
}

func deferredWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	use(ctx)
}

func deferredWithDeadline() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	use(ctx)
}

func deferredClosure() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() { cancel() }()
	use(ctx)
}

func deferredCause() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	use(ctx)
}

func deferredCauseClosure() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer func() { cancel(nil) }()
	use(ctx)
}

func deferredTimeoutCause() {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Second, nil)
	defer cancel()
	use(ctx)
}

func deferredAssignForm() {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	use(ctx)
}

func deferredAliasImport() {
	ctx, cancel := stdctx.WithCancel(context.Background())
	defer cancel()
	use(ctx)
}

func deferredRedeclared() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	use(ctx)
	next, cancel := context.WithCancel(ctx)
	defer cancel()
	use(next)
}

// ---- accepted: the cancel func escapes the function ----

func escapeReturned() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel
}

func escapeArgument() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	keep(cancel)
}

func escapeField() *holder {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	h := &holder{}
	h.cancel = cancel
	return h
}

func escapeCompositeLiteral() holder {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	return holder{cancel: cancel}
}

func escapeAssigned() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	other := cancel
	other()
}

func escapeMap() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	registry := map[string]context.CancelFunc{}
	registry["job"] = cancel
}

func escapeSlice() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	_ = []context.CancelFunc{cancel}
}

func escapeGoStatement() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		cancel()
	}()
}

func escapeGoCall() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	go cancel()
}

func escapeClosure() func() {
	ctx, cancel := context.WithCancel(context.Background())
	use(ctx)
	return func() { cancel() }
}

func escapePackageVariable() {
	var ctx context.Context
	ctx, stored = context.WithCancel(context.Background())
	use(ctx)
}

func escapeCapturedVariable() {
	var cancel context.CancelFunc
	run := func() {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		use(ctx)
	}
	run()
	keep(cancel)
}

// ---- accepted: a blank cancel is govet lostcancel's business ----

func blankCancel() {
	ctx, _ := context.WithCancel(context.Background())
	use(ctx)
}
