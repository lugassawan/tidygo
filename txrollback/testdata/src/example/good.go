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

// A closure that assigns a transaction declared by its enclosing function leaves
// the rollback to that function.
func outerManagesTx(ctx context.Context, db *sql.DB) error {
	var tx *sql.Tx
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	begin := func() (err error) {
		tx, err = db.BeginTx(ctx, nil)
		return err
	}
	return begin()
}

// A transaction per loop iteration is not reported: a deferred Rollback in the
// loop body would pile up until the function returns.
func perIterationTx(db *sql.DB, items []int) error {
	for range items {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

type unitOfWork struct{ tx *sql.Tx }

func (u *unitOfWork) current() *sql.Tx { return u.tx }

func txFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(unitOfWork{}).(*sql.Tx)
	return tx
}

// A getter hands back a transaction somebody else began and owns.
func getterIsNotABegin(ctx context.Context, u *unitOfWork) error {
	tx := txFromContext(ctx)
	if _, err := tx.ExecContext(ctx, "UPDATE t SET x = 1"); err != nil {
		return err
	}
	cur := u.current()
	_, err := cur.ExecContext(ctx, "UPDATE t SET y = 1")
	return err
}

// A conversion is not a Begin either.
func conversionIsNotABegin(p *sql.Tx) error {
	tx := (*sql.Tx)(p)
	return tx.Commit()
}

type cleaner interface{ Cleanup(func()) }

// The Rollback runs from a function literal that is registered rather than deferred.
func rollbackInCleanup(t cleaner, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	return tx.Commit()
}

// The Rollback is handed to the caller as a method value.
func rollbackAsMethodValue(db *sql.DB) (func() error, func() error, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}
	return tx.Commit, tx.Rollback, nil
}

// A function literal that rolls the transaction back counts as managing it, whether
// or not the literal is itself deferred.
func rollbackInNestedClosureCounts(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	release := func() {
		defer tx.Rollback()
	}
	release()
	return tx.Commit()
}
