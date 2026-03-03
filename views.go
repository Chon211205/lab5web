package main

import (
	"database/sql"
	"fmt"
	"net"
)

func showHome(conn net.Conn, db *sql.DB) {
	rows, err := db.Query(`
		SELECT s.id, s.name, s.current_episode, s.total_episodes,
			COALESCE(r.rating, 0) as rating
		FROM series s
		LEFT JOIN ratings r ON r.series_id = s.id
		ORDER BY s.id
	`)
	if err != nil {
		response := "HTTP/1.1 500 Internal Server Error\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"DB Query error: " + err.Error()
		conn.Write([]byte(response))
		return
	}
	defer rows.Close()

	html := "<html><head><title>Series</title></head><body>"
	html += "<h1 align='center'>My Series Tracker</h1>"
	html += "<div align='center'><a href='/create'>Agregar nueva serie</a></div><br>"
	html += "<table border='1' align='center'>"
	html += "<tr><th>Name</th><th>Current</th><th>Total</th><th>Progress</th><th>Rating (0-10)</th><th>Action</th><th>Delete</th></tr>"

	for rows.Next() {
		var id, current, total, rating int
		var name string

		if err := rows.Scan(&id, &name, &current, &total, &rating); err != nil {
			response := "HTTP/1.1 500 Internal Server Error\r\n" +
				"Content-Type: text/plain\r\n\r\n" +
				"Scan error: " + err.Error()
			conn.Write([]byte(response))
			return
		}

		progressPercent := 0
		if total > 0 {
			progressPercent = (current * 100) / total
		}

		html += fmt.Sprintf(
			"<tr>"+
				"<td>%s</td>"+
				"<td>%d</td>"+
				"<td>%d</td>"+
				"<td><progress value='%d' max='%d'></progress> %d%%</td>"+
				"<td>"+
				"<span id='rateText-%d'>%d</span> "+
				"<input type='range' min='0' max='10' value='%d' "+
				"oninput='updateRateText(%d, this.value)' "+
				"onchange='setRating(%d, this.value)'>"+
				"</td>"+
				"<td><button onclick='nextEpisode(%d)'>+1</button></td>"+
				"<td><button onclick='deleteSerie(%d)'>Eliminar</button></td>"+
				"</tr>",
			name,              // %s
			current,           // %d
			total,             // %d
			current,           // %d (progress value)
			total,             // %d (progress max)
			progressPercent,   // %d%%
			id,                // %d (rateText-id)
			rating,            // %d (texto rating)
			rating,            // %d (value del slider)
			id,                // %d (updateRateText)
			id,                // %d (setRating)
			id,                // %d (nextEpisode)
			id,                // %d (deleteSerie)
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
  await fetch(url, { method: "POST" })
  location.reload()
}

async function deleteSerie(id) {
  const url = "/series?id=" + id
  await fetch(url, { method: "DELETE" })
  location.reload()
}

function updateRateText(id, value) {
  const el = document.getElementById("rateText-" + id)
  if (el) el.textContent = value
}

async function setRating(seriesId, value) {
  const url = "/rate?series_id=" + seriesId + "&value=" + value
  await fetch(url, { method: "POST" })
}
`
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/javascript\r\n\r\n" +
		js
	conn.Write([]byte(response))
}