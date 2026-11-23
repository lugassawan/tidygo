package example

import (
	"context"
	"database/sql"
)

func commitOnly(db *sql.DB) error {
	tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO t VALUES (1)"); err != nil {
		return err
	}
	return tx.Commit()
}

func manualRollbackOnly(db *sql.DB) error {
	tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO t VALUES (1)"); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func beginTx(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil) // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}

func connBeginTx(ctx context.Context, conn *sql.Conn) error {
	t, err := conn.BeginTx(ctx, &sql.TxOptions{}) // want `transaction t has no deferred Rollback`
	if err != nil {
		return err
	}
	return t.Commit()
}

func valueSpec(db *sql.DB) error {
	var tx, err = db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}

func plainAssign(db *sql.DB) error {
	var tx *sql.Tx
	var err error
	tx, err = db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}

func viaWrapper(db *sql.DB) error {
	tx, err := begin(db) // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}

func secondTxUnprotected(db *sql.DB) error {
	a, err := db.Begin()
	if err != nil {
		return err
	}
	defer a.Rollback()

	b, err := db.Begin() // want `transaction b has no deferred Rollback`
	if err != nil {
		return err
	}
	if err := b.Commit(); err != nil {
		return err
	}
	return a.Commit()
}

func nestedClosureBeginsOwnTx(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	run := func() error {
		inner, err := db.Begin() // want `transaction inner has no deferred Rollback`
		if err != nil {
			return err
		}
		return inner.Commit()
	}
	if err := run(); err != nil {
		return err
	}
	return tx.Commit()
}

func shadowedInnerUnprotected(db *sql.DB, cond bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if cond {
		tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	return tx.Commit()
}

func shadowedOuterUnprotected(db *sql.DB, cond bool) error {
	tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	if cond {
		tx, err := db.Begin()
		if err == nil {
			defer tx.Rollback()
		}
	}
	return tx.Commit()
}

var literal = func(db *sql.DB) error {
	tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ---- a Begin inside a loop with no Rollback in the loop body ----

func rangeLoopLeaks(db *sql.DB, items []int) error {
	for range items {
		tx, err := db.Begin() // want `transaction tx begun in a loop has no Rollback`
		if err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE t SET x = 1"); err != nil {
			return err // leaks the transaction and its connection
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func threeClauseLoopLeaks(db *sql.DB, n int) error {
	for i := 0; i < n; i++ {
		tx, err := db.Begin() // want `transaction tx begun in a loop has no Rollback`
		if err != nil {
			return err
		}
		_ = tx.Commit()
	}
	return nil
}

func bareLoopLeaks(db *sql.DB) error {
	for {
		tx, err := db.Begin() // want `transaction tx begun in a loop has no Rollback`
		if err != nil {
			return err
		}
		return tx.Commit()
	}
}

func nestedLoopLeaks(db *sql.DB, rows [][]int) error {
	for _, row := range rows {
		for range row {
			tx, err := db.Begin() // want `transaction tx begun in a loop has no Rollback`
			if err != nil {
				return err
			}
			_ = tx.Commit()
		}
	}
	return nil
}

// A function literal is its own function: the loop is outside it, so the plain
// advice applies, and a defer inside the literal is the fix.
func literalInLoopLeaks(db *sql.DB, items []int) error {
	for range items {
		err := func() error {
			tx, err := db.Begin() // want `transaction tx has no deferred Rollback`
			if err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

// ---- a Start* wrapper is tracked like a Begin* one ----

func startTransaction(db *sql.DB) (*sql.Tx, error) {
	return db.Begin()
}

func viaStartWrapper(db *sql.DB) error {
	tx, err := startTransaction(db) // want `transaction tx has no deferred Rollback`
	if err != nil {
		return err
	}
	return tx.Commit()
}
