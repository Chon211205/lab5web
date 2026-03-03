package main

import (
	"database/sql"
	"fmt"
	"net"
)

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

	html += `<script src="/script.js"></script>`

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

func showScript(conn net.Conn) {
	js := `
async function nextEpisode(id) {
    const url = "/update?id=" + id
    const response = await fetch(url, { method: "POST" })
    location.reload()
}
`
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/javascript\r\n\r\n" +
		js

	conn.Write([]byte(response))
}