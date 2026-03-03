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

func setRating(db *sql.DB, seriesID string, rating string) {
	db.Exec(
		`INSERT INTO ratings (series_id, rating)
		 VALUES (?, ?)
		 ON CONFLICT(series_id) DO UPDATE SET rating = excluded.rating`,
		seriesID, rating,
	)
}