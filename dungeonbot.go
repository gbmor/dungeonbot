package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

// VERSION is set by make
var VERSION = ""

// HELPTEXT contains usage information
var HELPTEXT = []string{
	"!roll NdN[+-N]: roll dice or a die with optional modifier. Eg: !roll 1d20+4",
	"!add [campaign] $NAME: add a campaign entry called $NAME",
	"!append [campaign] $NAME $NOTE: append $NOTE to a campaign called $NAME",
	"!campaign $NAME: retrieve the campaign notes for $NAME",
}

// Config holds deserialized data from dungeonbot.yml
type Config struct {
	debug       bool
	nick        string
	user        string
	chans       []string
	server      string
	port        uint
	tls         bool
	pastebinURL string
}

func main() {
	if VERSION == "" {
		VERSION = "v0.1.0"
	}
	fmt.Println()
	fmt.Printf("\t-->  dungeonbot %s  <--\n", VERSION)
	fmt.Println("\tgithub.com/gbmor/dungeonbot")
	fmt.Println()

	conf := buildConf()
	host := fmt.Sprintf("%s:%d", conf.server, conf.port)

	conn := irc.IRC(conf.nick, conf.user)
	conn.VerboseCallbackHandler = false
	conn.Debug = conf.debug
	conn.UseTLS = conf.tls
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	db := &DB{}
	db.init("./dungeonbot.db")

	conn.AddCallback("001", func(e *irc.Event) {
		for i := 0; i < len(conf.chans); i++ {
			conn.Join(conf.chans[i])
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		splitRaw := strings.Split(e.Raw, " ")
		target := splitRaw[2]

		if strings.HasPrefix(e.Message(), "rain drop") {
			conn.Privmsg(target, "drop top")
			return
		}

		msg := strings.Split(e.Message(), " ")
		user := strings.ToLower(e.Nick)

		if len(msg) < 1 {
			return
		}

		if msg[0] == "!help" || msg[0] == "dungeonbot:" {
			for _, e := range HELPTEXT {
				conn.Privmsg(target, e)
			}
		}

		switch msg[0] {
		case "!roll":
			if len(msg) < 2 {
				conn.Privmsg(target, "Missing dice argument. Eg: !roll 1d20")
				break
			}

			out, err := parseDice(msg[1])
			if err != nil {
				conn.Privmsgf(target, "%s", err.Error())
				break
			}

			conn.Privmsgf(target, "%s", out)

		case "!campaign":
			if len(msg) < 2 {
				conn.Privmsgf(target, "Missing campaign name. Eg: !campaign gronkulousness")
				break
			}

			arg := strings.ToLower(strings.Join(msg[1:], " "))
			conn.Privmsgf(target, "Looking for %s campaign notes...", arg)

			raw, err := db.getCampaignNotes(arg)
			if err != nil {
				conn.Privmsgf(target, "No campaign notes for %s", arg)
				log.Printf("%s", err.Error())
				break
			}

			pbURL, err := pastebin(conf.pastebinURL, raw)
			if err != nil {
				conn.Privmsgf(target, "Error connecting to pastebin service")
				log.Printf("%s", err.Error())
				break
			}

			conn.Privmsgf(target, "%s", pbURL)

		case "!add":
			if len(msg) < 2 {
				conn.Privmsgf(target, "Missing subcommand: campaign|pc|npc|monster")
				break
			}
			if len(msg) < 3 {
				conn.Privmsgf(target, "Missing argument. Eg: !add campaign gronkulousness")
				break
			}
			subcommand := strings.ToLower(msg[1])
			name := strings.ToLower(msg[2])

			switch subcommand {
			case "campaign":
				if err := db.createCampaign(name, user); err != nil {
					conn.Privmsgf(target, "Error creating campaign")
					log.Printf("When creating campaign '%s': %s", msg[2], err.Error())
					break
				}
				conn.Privmsgf(target, "Campaign '%s' created", msg[2])
			case "pc":
			case "npc":
			case "monster":
			}

		case "!adduser":
			const argsError = "Missing arguments. Eg: !adduser [campaign|pc|npc|monster] gronkulousness somenerd"
			if len(msg) < 2 {
				conn.Privmsgf(target, argsError)
				break
			}
			if len(msg) < 3 {
				conn.Privmsgf(target, argsError)
				break
			}
			if len(msg) < 4 {
				conn.Privmsgf(target, argsError)
				break
			}

			subcommand := strings.ToLower(msg[1])
			name := strings.ToLower(msg[2])
			newuser := strings.ToLower(msg[3])

			switch subcommand {
			case "campaign":
				if err := db.addCampaignUser(name, user, newuser); err != nil {
					resp := ""
					if strings.Contains(err.Error(), "Not authorized") {
						resp = "Not authorized to modify user list"
					} else {
						resp = "Error adding user to user list"
					}
					conn.Privmsgf(target, resp)
					log.Printf("When adding user to campaign '%s': %s", user, err.Error())
					break
				}
				conn.Privmsgf(target, "User '%s' added to campaign '%s'", msg[3], msg[2])
			}
		case "!append":
			if len(msg) < 2 {
				conn.Privmsgf(target, "Missing subcommand: campaign|pc|npc|monster")
				break
			}
			if len(msg) < 3 {
				conn.Privmsgf(target, "Missing argument. Eg: !append campaign gronkulousness")
				break
			}
			if len(msg) < 4 {
				conn.Privmsgf(target, "Missing argument. Eg: !append campaign gronkulousness Don't trust the shopkeep in Grokuloustown")
				break
			}
			note := strings.Join(msg[3:], " ")
			subcommand := strings.ToLower(msg[1])
			name := strings.ToLower(msg[2])

			switch subcommand {
			case "campaign":
				if err := db.appendCampaign(name, note, user); err != nil {
					resp := ""
					if strings.Contains(err.Error(), "Not authorized") {
						resp = "not authorized to modify campaign notes"
					} else {
						resp = "Error appending note"
					}
					conn.Privmsgf(target, resp)
					log.Printf("When appending to notes for campaign '%s': %s", msg[2], err.Error())
					break
				}
				conn.Privmsgf(target, "Note appended to campaign '%s'", msg[2])
			}
		case "!clear":
		case "!delete":
		}
	})

	watchForInterrupt(conn, conf.nick, db.conn)

	if err := conn.Connect(host); err != nil {
		log.Fatalf("Error connecting: %s\n", err.Error())
	}

	conn.Loop()
}

func buildConf() Config {
	viper.SetConfigName("dungeonbot")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}

	chanWhole := viper.GetString("chans")
	chanSep := strings.Split(chanWhole, ",")

	return Config{
		debug:       viper.GetBool("debug_mode"),
		nick:        viper.GetString("nick"),
		user:        viper.GetString("user"),
		chans:       chanSep,
		server:      viper.GetString("server"),
		port:        viper.GetUint("port"),
		tls:         viper.GetBool("tls"),
		pastebinURL: viper.GetString("pastebin_url"),
	}
}

func watchForInterrupt(conn *irc.Connection, nick string, db *sql.DB) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("\n\nCaught %v\n", sigint)
			conn.SendRawf("QUIT /me yeet %s", nick)

			if err := db.Close(); err != nil {
				log.Printf("Error closing database connection: %s", err.Error())
			}

			time.Sleep(50 * time.Millisecond)
			os.Exit(1)
		}
	}()
}
