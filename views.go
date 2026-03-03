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

	html := "<html><head><title>Series</title>" +
		`<style>
			body{
				font-family: Arial, sans-serif;
				background: #e9910d;
				margin: 0;
				padding: 30px;
			}
			.container{
				max-width: 980px;
				margin: 0 auto;
				background: white;
				padding: 22px;
				border-radius: 12px;
				box-shadow: 0 6px 18px rgba(0,0,0,0.08);
			}
			h1{
				margin: 0 0 12px 0;
				text-align: center;
			}
			.topbar{
				display: flex;
				justify-content: center;
				margin-bottom: 18px;
			}
			a{
				text-decoration: none;
				color: #2b59ff;
				font-weight: 600;
			}
			table{
				width: 100%;
				border-collapse: collapse;
				overflow: hidden;
				border-radius: 10px;
			}
			th, td{
				padding: 12px 10px;
				border-bottom: 1px solid #e8e8ef;
				text-align: center;
			}
			th{
				background: #111827;
				color: white;
				font-weight: 700;
			}
			tr:nth-child(even){
				background: #f9fafb;
			}
			.btn{
				border: none;
				padding: 8px 12px;
				border-radius: 8px;
				cursor: pointer;
				font-weight: 700;
			}
			.btnPlus{
				background: #16a34a;
				color: white;
			}
			.btnDel{
				background: #dc2626;
				color: white;
			}
			progress{
				width: 160px;
				height: 14px;
			}
			input[type="range"]{
				width: 160px;
				vertical-align: middle;
			}
			@media (max-width: 720px){
				body{ padding: 14px; }
				progress{ width: 110px; }
				input[type="range"]{ width: 110px; }
				th, td{ padding: 10px 6px; font-size: 14px; }
			}
		</style>` +
		`</head><body>`

	html += "<div class='container'>"
	html += "<h1>Crunchyroll</h1>"
	html += "<div class='topbar'><a href='/create'>Agregar nueva serie</a></div>"
	html += "<table>"
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
				"<td><button class='btn btnPlus' onclick='nextEpisode(%d)'>+1</button></td>"+
				"<td><button class='btn btnDel' onclick='deleteSerie(%d)'>Eliminar</button></td>"+
				"</tr>",
			name,
			current,
			total,
			current,
			total,
			progressPercent,
			id,
			rating,
			rating,
			id,
			id,
			id,
			id,
		)
	}

	html += "</table>"
	html += `<script src="/script.js"></script>`
	html += "</div>"
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
<style>
	body{
		font-family: Arial, sans-serif;
		background: #ec8f04;
		margin: 0;
		padding: 30px;
	}
	.container{
		max-width: 520px;
		margin: 0 auto;
		background: white;
		padding: 22px;
		border-radius: 12px;
		box-shadow: 0 6px 18px rgba(0,0,0,0.08);
	}
	h1{
		margin: 0 0 14px 0;
		text-align: center;
	}
	label{
		font-weight: 700;
	}
	input{
		width: 100%;
		padding: 10px;
		border-radius: 10px;
		border: 1px solid #e5e7eb;
		margin-top: 6px;
	}
	.btn{
		width: 100%;
		margin-top: 14px;
		border: none;
		padding: 10px 12px;
		border-radius: 10px;
		cursor: pointer;
		font-weight: 800;
		background: #111827;
		color: white;
	}
	.back{
		text-align: center;
		margin-top: 14px;
	}
	a{
		text-decoration: none;
		color: #2b59ff;
		font-weight: 600;
	}
</style>
</head>
<body>
<div class="container">
<h1>Agregar Nueva Serie</h1>

<form method="POST" action="/create">
    <label>Nombre:</label>
    <input type="text" name="name" required>

    <label style="display:block; margin-top:12px;">Episodio actual:</label>
    <input type="number" name="current" required>

    <label style="display:block; margin-top:12px;">Total de episodios:</label>
    <input type="number" name="total" required>

    <button class="btn" type="submit">Guardar</button>
</form>

<div class="back">
<a href="/">Volver al inicio</a>
</div>
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
