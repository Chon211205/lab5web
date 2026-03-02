package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
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

	for {
		line, _ := reader.ReadString('\n')
		if line == "\r\n" {
			break
		}
	}

	if method != "GET" || path != "/" {
		return
	}

	rows, _ := db.Query("SELECT id, name, current_episode, total_episodes FROM series ORDER BY id ASC")
	defer rows.Close()

	html := `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>My Series Tracker</title>
  <style>
    body{
      font-family: Arial, sans-serif;
      background:#0f172a;
      color:#ffffff;
      padding: 24px;
      margin:0;
    }
    .light{
      background:#ffffff;
      color:#111827;
    }
    h1{
      text-align:center;
      margin: 0 0 14px 0;
    }
    .controls{
      display:flex;
      gap:10px;
      justify-content:center;
      flex-wrap:wrap;
      margin-bottom: 14px;
    }
    input{
      padding: 8px 10px;
      border: 1px solid #334155;
      border-radius: 6px;
      background: transparent;
      color: inherit;
      min-width: 220px;
      outline: none;
    }
    button{
      padding: 8px 10px;
      border: 1px solid #334155;
      border-radius: 6px;
      background: transparent;
      color: inherit;
      cursor:pointer;
    }
    button:hover{
      opacity: 0.9;
    }
    .stats{
      text-align:center;
      margin-bottom: 12px;
      opacity: 0.95;
    }
    table{
      border-collapse: collapse;
      width: 85%;
      margin: 0 auto;
      background: rgba(255,255,255,0.06);
    }
    .light table{
      background: #f3f4f6;
    }
    th, td{
      border:1px solid #334155;
      padding:10px;
      text-align:left;
    }
    th{
      background: rgba(255,255,255,0.10);
    }
    .light th{
      background:#e5e7eb;
    }
    tr.done{
      opacity: 0.85;
    }
    tr.almost{
      outline: 2px solid rgba(34,197,94,0.55);
      outline-offset: -2px;
    }
    .badge{
      display:inline-block;
      padding: 2px 8px;
      border-radius: 999px;
      border:1px solid #334155;
      font-size:12px;
      white-space:nowrap;
    }
  </style>
</head>
<body>
  <h1>My Series Tracker</h1>

  <div class="controls">
    <input id="search" type="text" placeholder="Buscar por nombre..." oninput="filterRows()" />
    <button onclick="sortByName()">Ordenar A-Z</button>
    <button onclick="sortByProgress()">Ordenar por progreso</button>
    <button onclick="toggleMode()">Modo claro/oscuro</button>
    <button onclick="toggleHideDone()" id="hideDoneBtn">Ocultar terminadas: OFF</button>
  </div>

  <div class="stats">
    <span class="badge" id="totalSeries">Series: 0</span>
    <span class="badge" id="totalWatched">Episodios vistos: 0</span>
    <span class="badge" id="totalEpisodes">Episodios totales: 0</span>
  </div>

  <table id="seriesTable" border="1" align="center">
    <tr>
      <th>#</th>
      <th>Name</th>
      <th>Current</th>
      <th>Total</th>
      <th>Progress %</th>
    </tr>
`

	for rows.Next() {
		var id, current, total int
		var name string

		rows.Scan(&id, &name, &current, &total)

		html += fmt.Sprintf(
			`<tr data-name="%s" data-current="%d" data-total="%d">
        <td>%d</td>
        <td>%s</td>
        <td>%d</td>
        <td>%d</td>
        <td class="progressCell"></td>
      </tr>`,
			escapeAttr(strings.ToLower(name)), current, total,
			id, escapeHTML(name), current, total,
		)
	}

	html += `
  </table>

  <script>
    let hideDone = false;

    function computeProgress() {
      const table = document.getElementById("seriesTable");
      const rows = Array.from(table.querySelectorAll("tr")).slice(1);

      let seriesCount = 0;
      let sumWatched = 0;
      let sumTotal = 0;

      rows.forEach(row => {
        const current = parseInt(row.dataset.current || "0", 10);
        const total = parseInt(row.dataset.total || "0", 10);
        const percent = total > 0 ? Math.round((current / total) * 100) : 0;

        const progressCell = row.querySelector(".progressCell");
        if (progressCell) progressCell.textContent = percent + "%";

        row.classList.remove("almost", "done");
        if (current >= total && total > 0) {
          row.classList.add("done");
        } else if (total > 0 && (current / total) >= 0.8) {
          row.classList.add("almost");
        }

        seriesCount += 1;
        sumWatched += current;
        sumTotal += total;
      });

      document.getElementById("totalSeries").textContent = "Series: " + seriesCount;
      document.getElementById("totalWatched").textContent = "Episodios vistos: " + sumWatched;
      document.getElementById("totalEpisodes").textContent = "Episodios totales: " + sumTotal;
    }

    function filterRows() {
      const q = document.getElementById("search").value.toLowerCase().trim();
      const table = document.getElementById("seriesTable");
      const rows = Array.from(table.querySelectorAll("tr")).slice(1);

      rows.forEach(row => {
        const name = (row.dataset.name || "");
        const isDone = row.classList.contains("done");

        const matches = name.includes(q);
        const showByDone = !(hideDone && isDone);

        row.style.display = (matches && showByDone) ? "" : "none";
      });
    }

    function sortByName() {
      const table = document.getElementById("seriesTable");
      const rows = Array.from(table.querySelectorAll("tr")).slice(1);

      rows.sort((a,b) => {
        const an = a.dataset.name || "";
        const bn = b.dataset.name || "";
        return an.localeCompare(bn);
      });

      rows.forEach(r => table.appendChild(r));
      computeProgress();
      filterRows();
    }

    function sortByProgress() {
      const table = document.getElementById("seriesTable");
      const rows = Array.from(table.querySelectorAll("tr")).slice(1);

      rows.sort((a,b) => {
        const ac = parseInt(a.dataset.current || "0", 10);
        const at = parseInt(a.dataset.total || "0", 10);
        const bc = parseInt(b.dataset.current || "0", 10);
        const bt = parseInt(b.dataset.total || "0", 10);

        const ap = at > 0 ? (ac / at) : 0;
        const bp = bt > 0 ? (bc / bt) : 0;

        return bp - ap; // de mayor a menor
      });

      rows.forEach(r => table.appendChild(r));
      computeProgress();
      filterRows();
    }

    function toggleMode() {
      document.body.classList.toggle("light");
    }

    function toggleHideDone() {
      hideDone = !hideDone;
      document.getElementById("hideDoneBtn").textContent = "Ocultar terminadas: " + (hideDone ? "ON" : "OFF");
      filterRows();
    }

    // Inicializar
    computeProgress();
    filterRows();
  </script>
</body>
</html>`

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		"\r\n" +
		html

	conn.Write([]byte(response))
}

func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

func escapeAttr(s string) string {
	// Para atributos HTML simples (data-*)
	// Evita comillas y caracteres especiales.
	replacer := strings.NewReplacer(
		`"`, "",
		"'", "",
		"<", "",
		">", "",
	)
	return replacer.Replace(s)
}