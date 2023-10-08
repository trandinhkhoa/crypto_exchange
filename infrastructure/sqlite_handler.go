package infrastructure

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
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
