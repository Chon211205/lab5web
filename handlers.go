package main

import (
	"bufio"
	"database/sql"
	"net"
	"net/url"
	"strconv"
	"strings"
)

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

	// Leer headers + Content-Length (como tu guía)
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

	// Separar ruta y query (para /update?id=..., etc.)
	partsPath := strings.SplitN(path, "?", 2)
	route := partsPath[0]

	// GET /script.js
	if method == "GET" && route == "/script.js" {
		showScript(conn)
		return
	}

	// GET /
	if method == "GET" && route == "/" {
		showHome(conn, db)
		return
	}

	// GET /create
	if method == "GET" && route == "/create" {
		showCreateForm(conn)
		return
	}

	// 2.3 — POST /update?id=...
	if method == "POST" && route == "/update" {
		var id string
		if len(partsPath) > 1 {
			params, _ := url.ParseQuery(partsPath[1])
			id = params.Get("id")
		}

		updateNextEpisode(db, id)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"ok"
		conn.Write([]byte(response))
		return
	}

	// POST /create
	if method == "POST" && route == "/create" {
		bodyBytes := make([]byte, contentLength)
		reader.Read(bodyBytes)
		body := string(bodyBytes)

		values, _ := url.ParseQuery(body)

		name := values.Get("name")
		currentEp := values.Get("current")
		totalEps := values.Get("total")

		insertSeries(db, name, currentEp, totalEps)

		// POST/Redirect/GET
		response := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n\r\n"
		conn.Write([]byte(response))
		return
	}
}