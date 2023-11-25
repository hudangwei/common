package clickhouse

import (
	"database/sql"

	qb "github.com/didi/gendry/builder"
	"github.com/jmoiron/sqlx"
)

func Count(db *sqlx.DB, table string, where map[string]interface{}) (int, error) {
	var total int
	sqlStr, args, err := qb.BuildSelect(table, where, []string{"count(1)"})
	if err != nil {
		return 0, err
	}
	err = db.Get(&total, sqlStr, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}

func GetOne(db *sqlx.DB, res interface{}, table string, where map[string]interface{}, selectField []string) error {
	sqlStr, args, err := qb.BuildSelect(table, where, selectField)
	if err != nil {
		return err
	}

	err = db.Get(res, sqlStr, args...)
	if err != nil {
		return err
	}
	// if err != nil && err != sql.ErrNoRows {
	// 	return err
	// }
	return nil
}

func GetAll(db *sqlx.DB, res interface{}, table string, where map[string]interface{}, selectField []string) error {
	sqlStr, args, err := qb.BuildSelect(table, where, selectField)
	if err != nil {
		return err
	}

	err = db.Select(res, sqlStr, args...)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}
