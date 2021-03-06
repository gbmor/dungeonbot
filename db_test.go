package main

import (
	"reflect"
	"strings"
	"testing"
)

const testDBLocation = ":memory:"

func uninitDB(db *DB) {
	err := db.conn.Close()
	if err != nil {
		panic(err)
	}
}

func Test_pastebin(t *testing.T) {
	go t.Run("pastebin", func(t *testing.T) {
		egress := "this is a test paste"
		ingress, err := pastebin("termbin.com:9999", egress)
		if err != nil {
			t.Error(err)
		}
		if !strings.HasPrefix(ingress, "https://termbin.com") {
			t.Errorf("Expected %s, got %s", egress, ingress)
		}
	})
}
func Test_DB_init(t *testing.T) {
	t.Run("db init", func(t *testing.T) {
		db := initDB(testDBLocation)
		defer uninitDB(db)

		_, err := db.conn.Exec("INSERT OR REPLACE INTO pcs (user, campaign, char, notes) VALUES(?, ?, ?, ?);", "foobat", "testCampaign", "testPlayer", "some notes")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		row := PCRow{}
		tmprow := db.conn.QueryRow("SELECT * FROM pcs WHERE campaign='testCampaign'")
		err = tmprow.Scan(&row.user, &row.campaign, &row.char, &row.notes)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		if row.user != "foobat" {
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
		db := initDB(testDBLocation)
		defer uninitDB(db)

		_, err := db.conn.Exec("INSERT OR REPLACE INTO campaigns (name, users, notes) VALUES(?, ?, ?)", "gronkulousness", "dungeonbot", "degronklified the dragon on 13 feb")
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
		db := initDB(testDBLocation)
		defer uninitDB(db)

		err := db.createCampaign("testcampaign", "dungeonbot")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		_, err = db.getCampaignNotes("testcampaign")
		if err != nil && !strings.Contains(err.Error(), "no campaign notes") {
			t.Errorf("%s", err.Error())
		}
	})
}

func Test_appendCampaign(t *testing.T) {
	t.Run("append campaign notes", func(t *testing.T) {
		db := initDB(testDBLocation)
		defer uninitDB(db)

		err := db.createCampaign("foocampaign", "dungeonbot")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		err = db.appendCampaign("foocampaign", "some notes that shouldn't work", "fakedungeonbot")
		if err == nil {
			t.Error("Allowed unauthed user to append campaign notes")
		}

		err = db.appendCampaign("foocampaign", "some notes go here", "dungeonbot")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		row := CampaignRow{}
		rrow := db.conn.QueryRow("SELECT * FROM campaigns WHERE name='foocampaign'")
		rrow.Scan(&row.name, &row.users, &row.notes)

		if row.notes != "some notes go here\n\n" {
			t.Errorf("Got \"%s\", expected \"some notes go here\"", row.notes)
		}
	})
}

func Test_addCampaignuser(t *testing.T) {
	t.Run("add campaign users", func(t *testing.T) {
		db := initDB(testDBLocation)
		defer uninitDB(db)

		err := db.createCampaign("gronkulousness", "dungeonbot")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		err = db.addCampaignUser("gronkulousness", "dungeonbot", "foouser")
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		err = db.addCampaignUser("gronkulousness", "dungeonbot", "foouser")
		if err == nil {
			t.Error("Able to add user twice")
		}

		row := CampaignRow{}
		rrow := db.conn.QueryRow("SELECT * FROM campaigns WHERE name='gronkulousness'")
		rrow.Scan(&row.name, &row.users, &row.notes)

		if !reflect.DeepEqual(row.users, "dungeonbot foouser") {
			t.Errorf("Incorrect user list: %s", row.users)
		}
	})
}
