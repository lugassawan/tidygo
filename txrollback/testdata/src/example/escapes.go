package example

import "database/sql"

type store struct {
	tx *sql.Tx
}

func commitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func rollbackTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

func returned(db *sql.DB) (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func returnedNamedResult(db *sql.DB) (tx *sql.Tx, err error) {
	tx, err = db.Begin()
	return
}

func passedToHelper(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	return commitTx(tx)
}

func deferredHelper(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)
	return tx.Commit()
}

func storedInField(db *sql.DB, s *store) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	s.tx = tx
	return nil
}

func storedInLiteral(db *sql.DB) (*store, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &store{tx: tx}, nil
}

func storedInMap(db *sql.DB, m map[string]*sql.Tx) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	m["k"] = tx
	return nil
}

func storedInSlice(db *sql.DB, all []*sql.Tx) ([]*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return append(all, tx), nil
}

func assignedToOther(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	other := tx
	return other.Commit()
}

func usedInGoStatement(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	go func() { _ = tx.Commit() }()
	return nil
}

func sentOnChannel(db *sql.DB, ch chan<- *sql.Tx) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ch <- tx
	return nil
}
