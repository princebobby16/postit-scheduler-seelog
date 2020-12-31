package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"scheduler-microservice/pkg/logs"
)

var Connection *sql.DB

func Connect() {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	Connection = db

	err = db.Ping()
	if err != nil {
		log.Printf("Unable to connect to database")
		panic(err)
	}

	logs.Log("Connected to Postgres DB successfully")
}

func Disconnect() {
	logs.Log("Attempting to disconnect from db....")
	err := Connection.Close()
	if err != nil {
		logs.Log(err)
	}
	logs.Log("Disconnected from db successfully...")
}
