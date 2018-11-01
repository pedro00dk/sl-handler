package database

import (
    "database/sql"

    _ "github.com/mattn/go-sqlite3"
)

var sqliteVersion = "sqlite3"
var pathDataBase = "./database.db"
var database, _ = sql.Open(sqliteVersion, pathDataBase)

func CreateSchema(){
	var qFunctionTable = "CREATE TABLE IF NOT EXISTS function (id INTEGER PRIMARY KEY, name TEXT, cpus INTEGER, memory INTEGER, code TEXT, pack TEXT)"
    statement, _ := database.Prepare(qFunctionTable)
	statement.Exec()
	
	var qMetricTable = "CREATE TABLE IF NOT EXISTS metric (id INTEGER PRIMARY KEY, function_id TEXT, duration INTEGER, status_code INTEGER, throttler, concurrentExecutions INTEGER)"
    statement, _ = database.Prepare(qMetricTable)
	statement.Exec()
}

func InsertFunction(name string, cpus, memory int, code string, pack string){
	statement, err := database.Prepare("INSERT INTO function (name, cpus, memory, code, pack) VALUES (?, ?, ?, ?, ?)")
	checkErr(err)
	_, err = statement.Exec(name, cpus, memory, code, pack)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Close(){
	database.Close();
}