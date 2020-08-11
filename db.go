package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds the database connection
type DB struct {
	conn *sql.DB
}

// PCRow holds a given row from the database
type PCRow struct {
	user     string
	campaign string
	char     string
	notes    string
}

// CampaignRow holds a given row from table campaigns
type CampaignRow struct {
	name  string
	users string
	notes string
}

// NPCRow holds a given row from table npcs
type NPCRow struct {
	name  string
	users string
	stats string
	notes string
}

// MonsterRow holds a given row from table monsters
type MonsterRow struct {
	name  string
	users string
	stats string
	notes string
}

func initDB(path string) *DB {
	db := &DB{}
	err := db.init(path)
	if err != nil {
		log.Fatalf("Could not initialize database: %s", err.Error())
	}
	return db
}

func (db *DB) init(path string) error {
	var err error
	if db.conn, err = sql.Open("sqlite3", path); err != nil {
		return fmt.Errorf("Failed to open database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("%s", err.Error())
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS pcs (
		user TEXT NOT NULL,
		campaign TEXT NOT NULL,
		char TEXT NOT NULL,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `pcs`: %w", err)
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS campaigns (
		name TEXT NOT NULL UNIQUE,
                users TEXT NOT NULL,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `campaigns`: %w", err)
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS npcs (
		name TEXT NOT NULL UNIQUE,
                users TEXT NOT NULL,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `npcs`: %w", err)
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS monsters (
		name TEXT NOT NULL UNIQUE,
                users TEXT NOT NULL,
		notes TEXT
	);`)
	if err != nil {
		return fmt.Errorf("Couldn't create-if-not-exists table `monsters`: %w", err)
	}

	return nil
}

func (db *DB) getCampaignNotes(campaign string) (string, error) {
	if err := db.conn.Ping(); err != nil {
		return "", fmt.Errorf("Couldn't ping database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return "", err
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("%s", err.Error())
			tx.Rollback()
		}
	}()

	row := tx.QueryRow("SELECT * FROM campaigns WHERE name=:campaign", sql.Named("campaign", campaign))
	if row == nil {
		return "", fmt.Errorf("Couldn't query row in table campaigns, campaign: %s", campaign)
	}

	crow := CampaignRow{}
	if err := row.Scan(&crow.name, &crow.users, &crow.notes); err != nil {
		return "", fmt.Errorf("Querying campaign notes: %w", err)
	}
	if crow.notes == "" {
		return "", fmt.Errorf("no campaign notes for '%s':\n\t%v\n%v", campaign, row, crow)
	}
	return crow.notes, nil
}

func (db *DB) createCampaign(name, user string) error {
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("Couldn't ping database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("Couldn't begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("%s", err.Error())
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("INSERT INTO campaigns (name, users, notes) VALUES(?, ?, ?)", name, user, "")
	if err != nil {
		return fmt.Errorf("Couldn't execute statement: %w", err)
	}

	return nil
}

func (db *DB) appendCampaign(name, note, user string) error {
	if name == "" || note == "" {
		return errors.New("invalid name or note")
	}
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("Couldn't ping database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("SELECT * FROM campaigns WHERE name=:name")
	if err != nil {
		tx.Rollback()
		return err
	}

	defer func() {
		stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Printf("%s", err.Error())
			tx.Rollback()
		}
	}()

	rowRaw := stmt.QueryRow(sql.Named("name", name))
	if rowRaw == nil {
		return fmt.Errorf("Couldn't retrieve campaign notes to append, campaign: %s", name)
	}

	row := CampaignRow{}
	rowRaw.Scan(&row.name, &row.users, &row.notes)

	if !strings.Contains(row.users, user) {
		return errors.New("Not authorized to modify campaign notes")
	}

	row.notes = fmt.Sprintf("%s%s\n\n", row.notes, note)

	_, err = tx.Exec("INSERT OR REPLACE INTO campaigns (name, users, notes) VALUES(?, ?, ?)", row.name, row.users, row.notes)
	if err != nil {
		return fmt.Errorf("Couldn't execute statement: %w", err)
	}

	return nil
}

func (db *DB) addCampaignUser(name, requser, user string) error {
	if name == "" || user == "" {
		return errors.New("Invalid campaign or user")
	}
	if strings.ContainsAny(user, " \t") {
		return errors.New("usernames cannot contain whitespace")
	}
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("Couldn't ping database: %w", err)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Commit(); err != nil {
			log.Printf("%s", err.Error())
			tx.Rollback()
		}
	}()

	rowRaw := tx.QueryRow("SELECT * FROM campaigns WHERE name=:name", name)
	if rowRaw == nil {
		return fmt.Errorf("Couldn't retrieve campaign row. Campaign: %s", name)
	}

	row := CampaignRow{}
	rowRaw.Scan(&row.name, &row.users, &row.notes)

	if !strings.Contains(row.users, requser) {
		return errors.New("Not authorized to modify campaign")
	}

	if strings.Contains(row.users, user) {
		return errors.New("User already authorized")
	}

	row.users += fmt.Sprintf(" %s", strings.TrimSpace(user))

	_, err = tx.Exec("INSERT OR REPLACE INTO campaigns (name, users, notes) VALUES(?, ?, ?)", row.name, row.users, row.notes)
	if err != nil {
		return err
	}

	return nil
}
