package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

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

func handleClient(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	requestLine, _ := reader.ReadString('\n')
	parts := strings.Split(requestLine, " ")

	if len(parts) < 2 {
		return
	}

	method := parts[0]
	path := parts[1]

	// Leer headers + Content-Length como pide la guía
	contentLength := 0
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(lengthStr)
		}
	}

	// GET /
	if method == "GET" && path == "/" {
		showHome(conn, db)
		return
	}

	// GET /create
	if method == "GET" && path == "/create" {
		showCreateForm(conn)
		return
	}

	// POST /create
	if method == "POST" && path == "/create" {
		// Leer cuerpo después de \r\n\r\n leyendo EXACTAMENTE Content-Length bytes
		bodyBytes := make([]byte, contentLength)
		reader.Read(bodyBytes)

		body := string(bodyBytes)

		// Parsear application/x-www-form-urlencoded
		values, _ := url.ParseQuery(body)

		// Nombres como en tu form actual: name, current, total
		name := values.Get("name")
		current := values.Get("current")
		total := values.Get("total")

		// ✅ 1.2: por ahora solo imprimir lo que viene
		fmt.Println("RAW BODY:", body)
		fmt.Println("Nombre:", name)
		fmt.Println("Episodio actual:", current)
		fmt.Println("Total episodios:", total)

		// Tu inserción la dejamos (si quieres solo imprimir, comenta esto)
		db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, current, total,
		)

		// Redirigir a la pagina principal
		response := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n\r\n"
		conn.Write([]byte(response))
		return
	}
}

func showHome(conn net.Conn, db *sql.DB) {
	rows, _ := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	defer rows.Close()

	html := "<html><head><title>Series</title></head><body>"
	html += "<h1 align='center'>My Series Tracker</h1>"
	html += "<div align='center'><a href='/create'>Agregar nueva serie</a></div><br>"
	html += "<table border='1' align='center'>"
	html += "<tr><th>Name</th><th>Current</th><th>Total</th></tr>"

	for rows.Next() {
		var id, current, total int
		var name string
		rows.Scan(&id, &name, &current, &total)

		html += fmt.Sprintf(
			"<tr><td>%s</td><td>%d</td><td>%d</td></tr>",
			name, current, total,
		)
	}

	html += "</table></body></html>"

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n\r\n" +
		html

	conn.Write([]byte(response))
}

func showCreateForm(conn net.Conn) {
	html := `
<html>
<head>
<title>Create Series</title>
</head>
<body>
<h1 align="center">Agregar Nueva Serie</h1>

<form method="POST" action="/create" align="center">
    <label>Nombre:</label><br>
    <input type="text" name="name" required><br><br>

    <label>Episodio actual:</label><br>
    <input type="number" name="current" required><br><br>

    <label>Total de episodios:</label><br>
    <input type="number" name="total" required><br><br>

    <button type="submit">Guardar</button>
</form>

<br>
<div align="center">
<a href="/">Volver al inicio</a>
</div>

</body>
</html>
`

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n\r\n" +
		html

	conn.Write([]byte(response))
}