package repositories

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// ErrDuplicate signals a unique-constraint violation.
//
// Without this, a second project using an existing slug surfaces as a raw driver
// error and the handler reports a 500 — telling the editor the server is broken
// when in fact their input needs one change.
var ErrDuplicate = errors.New("duplicate value")

// mysqlDuplicate is error 1062, ER_DUP_ENTRY.
const mysqlDuplicate = 1062

// asDuplicate converts a driver duplicate-key error into ErrDuplicate, and
// returns every other error unchanged.
func asDuplicate(err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicate {
		return ErrDuplicate
	}
	return err
}
