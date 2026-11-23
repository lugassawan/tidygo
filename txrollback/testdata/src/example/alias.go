package example

import dbsql "database/sql"

func aliasedImport(db *dbsql.DB) error {
	tx, err := db.Begin() // want `transaction tx is not rolled back on every path: add "defer tx.Rollback\(\)" after Begin`
	if err != nil {
		return err
	}
	return tx.Commit()
}

func aliasedImportDeferred(db *dbsql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}
