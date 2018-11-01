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

func DeleteFunction(id int){
	statement, err := database.Prepare("DELETE FROM function WHERE id=?")
	checkErr(err)

	_, err = statement.Exec(id)
	checkErr(err)
}

func SelectFunction(name string){
	statement, err := database.Prepare("SELECT FROM function WHERE name=?")
	checkErr(err)

	_, err = statement.Exec(name)
	checkErr(err)
}

type Function struct{
	id int
	name string
	cpus int
	memory int
	code string
	pack string
}

func SelectAllFunction() []Function{
	rows, err := database.Query("SELECT * FROM function")
	checkErr(err)
	var functionList = make([]Function, 0)
	
	for rows.Next() {
		function := Function{}
		err = rows.Scan(&function.id, &function.name, &function.cpus, &function.memory,&function.code, &function.pack)
		checkErr(err)
		functionList=append(functionList,function)
	}

	return functionList
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Close(){
	database.Close();
}