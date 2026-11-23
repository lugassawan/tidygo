package example

import (
	"context"
	"database/sql"
)

func begin(db *sql.DB) (*sql.Tx, error) {
	return db.Begin()
}

func deferRollback(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}

func deferRollbackBeginTx(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}

func deferClosure(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	return tx.Commit()
}

func deferLaterInFunction(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO t VALUES (1)"); err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}

func deferClosureWithIf(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	err = tx.Commit()
	committed = err == nil
	return err
}

func viaWrapperDeferred(db *sql.DB) error {
	tx, err := begin(db)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}

func nestedClosureProtected(db *sql.DB) error {
	run := func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		return tx.Commit()
	}
	return run()
}

func blankTarget(db *sql.DB) error {
	_, err := db.Begin()
	return err
}

type fakeTx struct{}

func newFakeTx() *fakeTx { return &fakeTx{} }

func notSQLTx() {
	t := newFakeTx()
	_ = t
}

func fieldTarget(db *sql.DB, s *store) error {
	var err error
	s.tx, err = db.Begin()
	return err
}
