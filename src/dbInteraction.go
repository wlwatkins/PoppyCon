package main

import (
  "database/sql"
  "log"

  _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type dbSensorRow struct {
  sensorType string
  sensorID string
  date int64
  valueFloat float64
}

type dbWaterRow struct {
  which int
  date int64
}

func initDB() *sql.DB {

  // os.Remove("sqlite-database.db") // I delete the file to avoid duplicated records.
  //                                   // SQLite is a file based database.
  //
	// log.Println("Creating sqlite-database.db...")
	// file, err := os.Create("sqlite-database.db") // Create SQLite file
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// file.Close()
	// log.Println("sqlite-database.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "../sqlite-database.db") // Open the created SQLite File

	// createTable(sqliteDatabase) // Create Database Tables

  return sqliteDatabase

}


func createTable(db *sql.DB) {
  var createTableSQL string
  var statement *sql.Stmt
  var err error

	createTableSQL = `CREATE TABLE sensors (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"sensorType" VARCHAR(200),
		"sensorID" VARCHAR(200),
		"date" INTEGER,
    "valueFloat" DECIMAL(10,5)
	  );` // SQL Statement for Create Table

	log.Println("Create table...")
	statement, err = db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements


  createTableSQL = `CREATE TABLE watering (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"date" INTEGER,
		"which" INTEGER
	  );` // SQL Statement for Create Table

	log.Println("Create table...")
	statement, err = db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements

}

func insertSensorDB(data dbSensorRow) {
	insertSQL := `INSERT INTO sensors(sensorType, sensorID, date, valueFloat) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
                                                   // This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(data.sensorType, data.sensorID, data.date, data.valueFloat)
  if err != nil {
		log.Fatalln(err.Error())
	}
}

func insertWaterDB(data dbWaterRow) {
	insertSQL := `INSERT INTO watering(date, which) VALUES (?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
                                                   // This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(data.date, data.which)
  if err != nil {
		log.Fatalln(err.Error())
	}
}
