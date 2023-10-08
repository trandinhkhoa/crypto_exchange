package infrastructure

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func SetupDatabase(dbLocation string) *sql.DB {
	// Open SQLite database
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		logrus.Fatal("Unable to open database", err)
		// TODO: how to write defer only once for all err
		defer db.Close()
	}

	// Create table
	// TODO: avoid hardcoding all currencies
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"userid" TEXT PRIMARY KEY,
		"ETH" FLOAT,
		"USD" FLOAT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		logrus.Fatal("Unable to create table", err)
		defer db.Close()
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS buyOrders (
		"id" INTEGER PRIMARY KEY,
		"userid" TEXT,
		"size" INTEGER,
		"price" INTEGER,
		"timestamp" INTEGER
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		logrus.Fatal("Unable to create table", err)
		defer db.Close()
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS sellOrders (
		"id" INTEGER PRIMARY KEY,
		"userid" TEXT,
		"size" INTEGER,
		"price" INTEGER,
		"timestamp" INTEGER
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		logrus.Fatal("Unable to create table", err)
		defer db.Close()
	}

	logrus.Info("Connected to the database")
	return db
}
