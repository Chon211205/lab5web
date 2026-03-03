package main

import "database/sql"

func insertSeries(db *sql.DB, name string, currentEp string, totalEps string) {
	db.Exec(
		"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
		name, currentEp, totalEps,
	)
}

func updateNextEpisode(db *sql.DB, id string) {
	db.Exec(
		`UPDATE series
         SET current_episode = current_episode + 1
         WHERE id = ? AND current_episode < total_episodes`,
		id,
	)
}

func deleteSeries(db *sql.DB, id string) {
	db.Exec("DELETE FROM series WHERE id = ?", id)
}