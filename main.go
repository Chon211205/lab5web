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

	//POST /next?id=3  (incrementa episodio)
	if method == "POST" && strings.HasPrefix(path, "/next") {
		// path viene como: /next?id=3
		partsPath := strings.Split(path, "?")
		if len(partsPath) < 2 {
			return
		}

		queryStr := partsPath[1]
		q, _ := url.ParseQuery(queryStr)
		id := q.Get("id")

		db.Exec("UPDATE series SET current_episode = current_episode + 1 WHERE id = ?", id)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"OK"
		conn.Write([]byte(response))
		return
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
		// Leer cuerpo leyendo content-Length bytes
		bodyBytes := make([]byte, contentLength)
		reader.Read(bodyBytes)

		body := string(bodyBytes)

		// Parsear application
		values, _ := url.ParseQuery(body)

		// Parametros de la tabla
		name := values.Get("name")
		currentEp := values.Get("current")
		totalEps := values.Get("total")

		fmt.Println("RAW BODY:", body)
		fmt.Println("Nombre:", name)
		fmt.Println("Episodio actual:", currentEp)
		fmt.Println("Total episodios:", totalEps)

		// Insertar en la base de datos y redirigir a la pagina principal
		db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, currentEp, totalEps,
		)

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
	html += "<tr><th>Name</th><th>Current</th><th>Total</th><th>Action</th></tr>"

	for rows.Next() {
		var id, current, total int
		var name string
		rows.Scan(&id, &name, &current, &total)

		html += fmt.Sprintf(
			"<tr><td>%s</td><td>%d</td><td>%d</td><td><button onclick='nextEpisode(%d)'>+1</button></td></tr>",
			name, current, total, id,
		)
	}

	html += "</table>"

	//llamar POST /next?id=ID y recargar
	html += `
<script>
function nextEpisode(id){
	fetch("/next?id=" + id, { method: "POST" })
	.then(() => window.location.reload());
}
</script>
`

	html += "</body></html>"

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