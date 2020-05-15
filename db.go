package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds the database connection
type DB struct {
	mu   sync.RWMutex
	conn *sql.DB
}

// PCRow holds a given row from the database
type PCRow struct {
	nick     string
	campaign string
	char     string
	notes    string
}

// CampaignRow holds a given row from table campaigns
type CampaignRow struct {
	name  string
	notes string
}

// NPCRow holds a given row from table npcs
type NPCRow struct {
	name  string
	stats string
	notes string
}

// MonsterRow holds a given row from table monsters
type MonsterRow struct {
	name  string
	stats string
	notes string
}

func pastebin(pastebin string, input string) (string, error) {
	pbconn, err := net.Dial("tcp", pastebin)
	if err != nil {
		return "", fmt.Errorf("Error connecting to pastebin service: %w", err)
	}
	defer pbconn.Close()

	if _, err := pbconn.Write([]byte(input)); err != nil {
		return "", fmt.Errorf("Error sending data to pastebin service: %w", err)
	}

	pbRdr := bufio.NewReader(pbconn)
	pbBytes, _, err := pbRdr.ReadLine()
	if err != nil {
		return "", fmt.Errorf("Error reading response from pastebin service: %w", err)
	}

	return string(pbBytes), err
}

func (db *DB) init(path string) error {
	var err error
	if db.conn, err = sql.Open("sqlite3", path); err != nil {
		return fmt.Errorf("Failed to open database: %w", err)
	}

	if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS pcs (
		nick TEXT NOT NULL,
		campaign TEXT NOT NULL,
		char TEXT NOT NULL,
		notes TEXT
	);`); err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `pcs`: %w", err)
	}

	if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS campaigns (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`); err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `campaigns`: %w", err)
	}

	if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS npcs (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`); err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `npcs`: %w", err)
	}

	if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS monsters (
		name TEXT NOT NULL UNIQUE,
		notes TEXT
	);`); err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `monsters`: %w", err)
	}

	return nil
}

func (db *DB) getCampaignNotes(campaign string) (string, error) {
	if err := db.conn.Ping(); err != nil {
		return "", fmt.Errorf("Couldn't ping database: %w", err)
	}

	row := db.conn.QueryRow("SELECT * FROM campaigns WHERE name=:campaign", sql.Named("campaign", campaign))
	if row == nil {
		return "", fmt.Errorf("Couldn't query row in table campaigns, campaign: %s", campaign)
	}

	crow := CampaignRow{}
	err := row.Scan(&crow.name, &crow.notes)
	if err != nil {
		return "", fmt.Errorf("Querying campaign notes: %w", err)
	}
	if crow.notes == "" {
		return "", fmt.Errorf("No campaign notes for '%s':\n\t%v\n%v\n", campaign, row, crow)
	}
	return crow.notes, nil
}

func (db *DB) createCampaign(name string) error {
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("Couldn't ping database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("Couldn't begin transaction: %w", err)
	}

	_, err = tx.Exec("INSERT INTO campaigns (name, notes) VALUES(?, ?)", name, "")
	if err != nil {
		return fmt.Errorf("Couldn't execute statement: %w", err)
	}

	return tx.Commit()
}

func (db *DB) appendCampaign(name, note string) error {
	if name == "" || note == "" {
		return errors.New("invalid name or note")
	}
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("Couldn't ping database: %w", err)
	}

	rowRaw := db.conn.QueryRow("SELECT * FROM campaigns WHERE name=:name", name)
	if rowRaw == nil {
		return fmt.Errorf("Couldn't retrieve campaign notes to append, campaign: %s", name)
	}

	row := &CampaignRow{}
	rowRaw.Scan(&row.name, &row.notes)

	row.notes = fmt.Sprintf("%s%s\n\n", row.notes, note)

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("Couldn't begin transaction: %w", err)
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO campaigns (name, notes) VALUES(?, ?)", row.name, row.notes)
	if err != nil {
		return fmt.Errorf("Couldn't execute statement: %w", err)
	}

	return tx.Commit()
}
