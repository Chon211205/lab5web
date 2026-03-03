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

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return
	}

	method := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])
	path = strings.TrimRight(path, "\r\n")

	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(lengthStr)
		}
	}

	partsPath := strings.SplitN(path, "?", 2)
	route := strings.TrimSpace(partsPath[0])

	if method == "GET" && route == "/script.js" {
		showScript(conn)
		return
	}

	if method == "GET" && route == "/" {
		showHome(conn, db)
		return
	}

	if method == "GET" && route == "/create" {
		showCreateForm(conn)
		return
	}

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

	if method == "DELETE" && route == "/series" {
		var id string
		if len(partsPath) > 1 {
			params, _ := url.ParseQuery(partsPath[1])
			id = params.Get("id")
		}

		deleteSeries(db, id)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"ok"
		conn.Write([]byte(response))
		return
	}

	if method == "POST" && route == "/create" {
		bodyBytes := make([]byte, contentLength)
		reader.Read(bodyBytes)
		body := string(bodyBytes)

		values, _ := url.ParseQuery(body)

		name := values.Get("name")
		currentEp := values.Get("current")
		totalEps := values.Get("total")

		insertSeries(db, name, currentEp, totalEps)

		response := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n\r\n"
		conn.Write([]byte(response))
		return
	}

	if method == "POST" && route == "/rate" {
		var seriesID, value string
		if len(partsPath) > 1 {
			params, _ := url.ParseQuery(partsPath[1])
			seriesID = params.Get("series_id")
			value = params.Get("value")
		}

		setRating(db, seriesID, value)

		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"ok"
		conn.Write([]byte(response))
		return
	}

	response := "HTTP/1.1 404 Not Found\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"not found"
	conn.Write([]byte(response))
}