package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func main(){

	var user string
	db1, err := sql.Open("sqlite3", "/home/isaac/DBs/live.db")
	if err != nil {
		fmt.Println(err)
	}
	db2, err := sql.Open("sqlite3", "/home/isaac/DBs/live.db")
	if err != nil {
		fmt.Println(err)
	}
	row := db1.QueryRow("SELECT username FROM encoders WHERE streamname = 'mobile'").Scan(&user)
	if row != nil {
		fmt.Println("error1", row)
	}
	fmt.Printf("1º Consulta: %s\n", user)
	row = db2.QueryRow("SELECT username FROM encoders WHERE streamname = 'mobile'").Scan(&user)
	if row != nil {
		fmt.Println("error2",row)
	}
	fmt.Printf("2º Consulta: %s\n", user)
	db1.Close() //cierro la primera queryrow
	
	row = db2.QueryRow("SELECT username FROM encoders WHERE streamname = 'mobile'").Scan(&user)
	if row != nil {
		fmt.Println("error3",row)
	}
	fmt.Printf("3º Consulta: %s\n", user)
	db2.Close() //cierro la segunda queryrow
}

