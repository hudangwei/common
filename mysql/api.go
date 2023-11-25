package mysql

import (
	"database/sql"
	"errors"

	qb "github.com/didi/gendry/builder"
	"github.com/jmoiron/sqlx"
)

var ErrOptimisticLock = errors.New("Optimistic Lock Error")

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

func GetList(db *sqlx.DB, res interface{}, table string, where map[string]interface{}, selectField []string, page, pageSize uint) (int, error) {
	total, err := Count(db, table, where)
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	if page != 0 {
		page = page - 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	where["_limit"] = []uint{page * pageSize, pageSize}
	sql, args, err := qb.BuildSelect(table, where, selectField)
	if err != nil {
		return total, err
	}
	if err = db.Select(res, sql, args...); err != nil {
		return total, err
	}
	return total, nil
}

func Save(db *sqlx.DB, table string, data map[string]interface{}) (int64, error) {
	sql, args, err := qb.BuildInsert(table, []map[string]interface{}{data})
	if err != nil {
		return 0, err
	}
	result, err := db.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func Update(db *sqlx.DB, table string, where, data map[string]interface{}) error {
	sql, args, err := qb.BuildUpdate(table, where, data)
	if err != nil {
		return err
	}
	res, err := db.Exec(sql, args...)
	if err != nil {
		return err
	}
	effected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if effected == 0 {
		return ErrOptimisticLock
	}
	return nil
}

func Delete(db *sqlx.DB, table string, where map[string]interface{}) error {
	sql, args, err := qb.BuildDelete(table, where)
	if err != nil {
		return err
	}
	_, err = db.Exec(sql, args...)
	return err
}
