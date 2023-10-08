package infrastructure

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/trandinhkhoa/crypto-exchange/controllers"
)

type SqliteDbHandler struct {
	dbConn *sql.DB
}

func NewSqliteDbHandler(dbFileName string) *SqliteDbHandler {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		db.Close()
		panic(fmt.Sprintf("Unable to open database %s", err.Error()))
	}

	sqliteDbHandler := &SqliteDbHandler{
		dbConn: db,
	}

	logrus.Info("Connected to the database")

	return sqliteDbHandler
}

func (sqlDbHandler *SqliteDbHandler) Exec(statement string) error {
	retryCount := 4
	var err error
	if _, err = sqlDbHandler.dbConn.Exec(statement); err != nil {
		logrus.Error(fmt.Sprintf("Unable to exec statement. Error: %s. Retrying... Statement: %s", err.Error(), statement))
		for retryCount > 0 {
			_, err := sqlDbHandler.dbConn.Exec(statement)
			if err == nil {
				logrus.Info(fmt.Sprintf("Retry OK. Statement: %s", statement))
				break
			}
			retryCount -= 1
		}
	}
	if err != nil {
		logrus.Error(fmt.Sprintf("Unable to exec statement. Error: %s. Statement: %s", err.Error(), statement))
		return err
	} else {
		return nil
	}
}

func (sqlDbHandler *SqliteDbHandler) Close() {
	sqlDbHandler.dbConn.Close()
}

type SqliteRow struct {
	Rows *sql.Rows
}

func (r SqliteRow) Scan(dest ...interface{}) {
	if err := r.Rows.Scan(dest...); err != nil {
		logrus.Error(err)
	}
}

func (r SqliteRow) Next() bool {
	return r.Rows.Next()
}

func (sqlDbHandler *SqliteDbHandler) Query(statement string) controllers.Row {
	// Step 2: Execute SQL SELECT query
	rows, err := sqlDbHandler.dbConn.Query(statement)
	if err != nil {
		logrus.Error(err)
		return new(SqliteRow)
	}
	defer rows.Close()
	row := new(SqliteRow)
	row.Rows = rows
	return row
}
