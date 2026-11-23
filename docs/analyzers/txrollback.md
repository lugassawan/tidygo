# txrollback

Requires a `*sql.Tx` to have a `defer tx.Rollback()` in the function that begins it. A transaction that returns early without committing or rolling back keeps its connection and any locks until the pool times out. `Rollback` after a successful `Commit` is a harmless no-op that returns `sql.ErrTxDone`, so deferring it right after `Begin` is always safe.

The analyzer needs type information. It tracks any assignment or `var` whose left-hand variable is a `*database/sql.Tx` and whose right-hand side calls a function or method named `Begin…` or `Start…` (case-insensitive), which covers `DB.Begin`, `DB.BeginTx`, `Conn.BeginTx` and wrappers such as `beginTx(db)` or `s.startTransaction(ctx)`. The deferral may be anywhere in the same function, either `defer tx.Rollback()` or a deferred function literal that calls `tx.Rollback()` (including `_ = tx.Rollback()` or inside an `if`). A manual `tx.Rollback()` in an error branch does not count, because it misses the other return paths and panics. A `Begin` inside a `for` or `range` body is checked differently, because a deferred `Rollback` there would pile up until the function returns. It is accepted when the loop body calls `tx.Rollback()` on that variable anywhere (deferred or not), and otherwise reported with the advice to extract the iteration into a function that has its own `defer tx.Rollback()`.

```go
// Bad
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
if _, err := tx.ExecContext(ctx, q); err != nil {
	return err // leaks the transaction
}
return tx.Commit()

// Good
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
defer tx.Rollback()
```

A transaction that escapes the function is skipped, because the caller owns it: when it is returned, passed as an argument (including `defer cleanup(tx)`), stored in a field or composite literal, assigned to another variable, or used in a `go` statement. It is also treated as managed when its `Rollback` is handed to code that runs later: called inside any function literal (`t.Cleanup(func() { _ = tx.Rollback() })`), or passed on as a method value (`return tx.Commit, tx.Rollback, nil`).

## Known limitations

- Only `*database/sql.Tx` is tracked; other transaction types (`pgx.Tx`, `sqlx.Tx`) are not.
- The check does not prove a `Rollback` runs on *every* path, only that one is deferred in the function.
- Passing the transaction to any function counts as handing over ownership, so the common `tx, _ := db.Begin(); repo.insert(tx, row); tx.Commit()` is not reported. Methods called on `tx` itself (`tx.Exec`, `tx.Commit`) are not hand-overs.
- A call starts a transaction only when its name begins with `Begin` or `Start`. A getter such as `tx := txFromContext(ctx)` and a type conversion are not reported, but a wrapper with another name (`openTx`, `newUnitOfWork`) is not tracked either, and neither is a call through a function value.
- A variable assigned from `Begin` twice in one function is reported at each assignment.
- In a loop, any `Rollback` call on the variable in the loop body is accepted, even one that covers only some of the exits, so an iteration that rolls back on one error path and returns early on another is not reported.
