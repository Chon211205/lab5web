package main

import (
	"database/sql"
	"fmt"
	"net"

	_ "modernc.org/sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", "file:series.db")
	defer db.Close()

	listener, _ := net.Listen("tcp", ":8080")
	fmt.Println("Servidor corriendo en http://localhost:8080")

	for {
		conn, _ := listener.Accept()
		go handleClient(conn, db)
	}
}
