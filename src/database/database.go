package database

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

func CreateSchema(){
	var sqliteVersion = "sqlite3"
	var pathDataBase = "./database.db"
	var qFunctionTable = "CREATE TABLE IF NOT EXISTS function (id INTEGER PRIMARY KEY, name TEXT, cpus INTEGER, memory INTEGER, code TEXT, package TEXT)"
	database, _ := sql.Open(sqliteVersion, pathDataBase)
    statement, _ := database.Prepare(qFunctionTable)
	statement.Exec()
	
	var qMetricTable = "CREATE TABLE IF NOT EXISTS metric (id INTEGER PRIMARY KEY, function_id TEXT, duration INTEGER, status_code INTEGER, throttler, concurrentExecutions INTEGER)"
    statement, _ = database.Prepare(qMetricTable)
	statement.Exec()
}

func InsertFunction(){

}