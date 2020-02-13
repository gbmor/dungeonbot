package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds the database connection
type DB struct {
	mu   *sync.RWMutex
	conn *sql.DB
}

// Row holds a given row from the database
type Row struct {
	nick     string `db:"nick"`
	campaign string `db:"campaignName"`
	char     string `db:"charName"`
	notes    string `db:"notes"`
}

func (db *DB) init() error {
	var err error
	db.conn, err = sql.Open("sqlite3", "./dungeonbot.db")
	if err != nil {
		return fmt.Errorf("Failed to open database: %s", err.Error())
	}

	_, err = db.conn.Exec(`CREATE TABLE IF NOT EXISTS pcs (
		nick TEXT NOT NULL,
		campaign TEXT NOT NULL,
		char TEXT NOT NULL,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `pcs`")
	}

	_, err = db.conn.Exec(`CREATE TABLE IF NOT EXISTS campaigns (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `campaigns`")
	}

	_, err = db.conn.Exec(`CREATE TABLE IF NOT EXISTS npcs (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `npcs`'")
	}

	_, err = db.conn.Exec(`CREATE TABLE IF NOT EXISTS monsters (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `monsters`")
	}

	db.mu = &sync.RWMutex{}
	return nil
}
