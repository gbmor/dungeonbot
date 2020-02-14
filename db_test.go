package main

import (
	"strings"
	"testing"
)

func Test_pastebin(t *testing.T) {
	go t.Run("pastebin", func(t *testing.T) {
		egress := "this is a test paste"
		ingress, err := pastebin("termbin.com:9999", egress)
		if err != nil {
			t.Error(err)
		}
		if !strings.HasPrefix(ingress, "https://termbin.com/") {
			t.Errorf("Expected %s, got %s", egress, ingress)
		}
	})
}
func Test_DB_init(t *testing.T) {
	t.Run("db init", func(t *testing.T) {
		db := &DB{}
		err := db.init()
		if err != nil || db.conn == nil {
			t.Errorf("%s", err.Error())
		}
		defer db.conn.Close()

		_, err = db.conn.Exec("INSERT OR REPLACE INTO pcs (nick, campaign, char, notes) VALUES(?, ?, ?, ?);", "foobat", "testCampaign", "testPlayer", "some notes")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		row := PCRow{}
		tmprow := db.conn.QueryRow("SELECT * FROM pcs WHERE campaign='testCampaign'")
		err = tmprow.Scan(&row.nick, &row.campaign, &row.char, &row.notes)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		if row.nick != "foobat" {
			t.Errorf("Did not retrieve nick name")
		}
		if row.campaign != "testCampaign" {
			t.Errorf("Did not retrieve campaign name")
		}
		if row.char != "testPlayer" {
			t.Errorf("Did not retrieve player name")
		}
		if row.notes != "some notes" {
			t.Errorf("Did not retrieve notes")
		}
	})
}

func Test_getCampaignNotes(t *testing.T) {
	t.Run("get campaign notes", func(t *testing.T) {
		db := &DB{}
		err := db.init()
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer db.conn.Close()

		_, err = db.conn.Exec("INSERT OR REPLACE INTO campaigns (name, notes) VALUES(?, ?)", "gronkulousness", "degronklified the dragon on 13 feb")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		out, err := db.getCampaignNotes("gronkulousness")
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		if len(out) == 0 || out != "degronklified the dragon on 13 feb" {
			t.Errorf("Output: %s", out)
		}
	})
}

func Test_createCampaign(t *testing.T) {
	t.Run("create campaign entry", func(t *testing.T) {
		db := &DB{}
		err := db.init()
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer db.conn.Close()

		err = db.createCampaign("testcampaign")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		_, err = db.getCampaignNotes("testcampaign")
		if err != nil {
			t.Errorf("%s", err.Error())
		}
	})
}
