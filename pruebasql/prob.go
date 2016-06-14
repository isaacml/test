package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
	"time"
	"fmt"
)

var (
	db      *sql.DB
	db_mu	sync.Mutex
)

// Inicializamos la conexion a BD y el log de errores
func init() {
	var err_db error
	db, err_db = sql.Open("sqlite3", "/var/segments/probsql.db")
	if err_db != nil {
		log.Println(err_db)
	}
}

func main(){
	
	tiempo := time.Now()
	entero:=""
	for i:=0; i <= 1000; i++{
		query, err := db.Query("SELECT filename FROM `SEGMENTOS` WHERE filename = 2")
		if err != nil {
			log.Println(err)
		}
		for query.Next() {
		err = query.Scan(&entero)
		if err != nil {
			log.Println(err)
		}
	}
	}
	
	
	duracion := time.Since(tiempo).Nanoseconds()
	fmt.Printf("%d  -  %s\n",duracion/1000000,entero)
	db.Close()	
}

