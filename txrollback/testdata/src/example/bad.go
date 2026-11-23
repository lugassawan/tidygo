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
