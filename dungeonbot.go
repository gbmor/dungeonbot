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

var CACHE = &notesCache{
	kv: make(map[string]string),
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
	dbLocation  string
	signOff     string
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
	CACHE.pb = conf.pastebinURL
	helpText := genHelpText(conf)

	conn := irc.IRC(conf.nick, conf.user)
	conn.VerboseCallbackHandler = false
	conn.Debug = conf.debug
	conn.UseTLS = conf.tls
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	db := initDB(conf.dbLocation)

	conn.AddCallback("001", func(e *irc.Event) {
		for i := 0; i < len(conf.chans); i++ {
			conn.Join(conf.chans[i])
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		target := e.Arguments[0]

		if strings.HasPrefix(e.Message(), "rain drop") {
			conn.Privmsg(target, "drop top")
			return
		}

		msg := strings.Split(e.Message(), " ")
		user := strings.ToLower(e.Nick)

		if len(msg) < 1 {
			return
		}

		if msg[0] == "!help" || msg[0] == "dungeonbot:" || msg[0] == "!botlist" {
			conn.Privmsg(target, helpText)
		}

		switch msg[0] {
		case "!roll":
			if len(msg) < 2 {
				conn.Privmsg(target, "Missing dice argument. Eg: !roll 1d20")
				break
			}

			out, err := parseDice(msg[1])
			if err != nil {
				conn.Privmsg(target, err.Error())
				break
			}

			conn.Privmsg(target, out)

		case "!campaign":
			if len(msg) < 2 {
				conn.Privmsg(target, "Missing campaign name. Eg: !campaign gronkulousness")
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

			//pbURL, err := pastebin(conf.pastebinURL, raw)
			//if err != nil {
			//	conn.Privmsg(target, "Error connecting to pastebin service")
			//	log.Printf("%s", err.Error())
			//	break
			//}

			pbURL := CACHE.bap(string(raw))
			conn.Privmsg(target, pbURL)

		case "!add":
			const argsError = "Incorrect arguments. Eg: !add [campaign|monster|npc|pc] gronkulousness"
			if len(msg) < 3 {
				conn.Privmsg(target, argsError)
				break
			}
			subcommand := strings.ToLower(msg[1])
			name := strings.ToLower(msg[2])

			switch subcommand {
			case "campaign":
				if err := db.createCampaign(name, user); err != nil {
					conn.Privmsg(target, "Error creating campaign")
					log.Printf("When creating campaign '%s': %s", msg[2], err.Error())
					break
				}
				conn.Privmsgf(target, "Campaign '%s' created", msg[2])
			case "monster":
				conn.Privmsg(target, "unimplemented")
			case "pc":
				conn.Privmsg(target, "unimplemented")
			case "npc":
				conn.Privmsg(target, "unimplemented")
			default:
				conn.Privmsg(target, argsError)
			}

		case "!adduser":
			const argsError = "Incorrect arguments. Eg: !adduser [campaign|monster|npc|pc] gronkulousness somenerd"
			if len(msg) < 4 {
				conn.Privmsg(target, argsError)
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
					conn.Privmsg(target, resp)
					log.Printf("When adding user to campaign '%s': %s", user, err.Error())
					break
				}
				conn.Privmsgf(target, "User '%s' added to campaign '%s'", msg[3], msg[2])
			case "monster":
				conn.Privmsg(target, "unimplemented")
			case "npc":
				conn.Privmsg(target, "unimplemented")
			case "pc":
				conn.Privmsg(target, "unimplemented")
			default:
				conn.Privmsg(target, argsError)
			}
		case "!append":
			const argsError = "Incorrect arguments. Eg: !append [campaign|monster|npc|pc] gronkulousness The saxophone is a mimic"
			if len(msg) < 4 {
				conn.Privmsg(target, argsError)
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
					conn.Privmsg(target, resp)
					log.Printf("When appending to notes for campaign '%s': %s", msg[2], err.Error())
					break
				}
				conn.Privmsgf(target, "Note appended to campaign '%s'", msg[2])
			case "monster":
				conn.Privmsg(target, "unimplemented")
			case "pc":
				conn.Privmsg(target, "unimplemented")
			case "npc":
				conn.Privmsg(target, "unimplemented")
			default:
				conn.Privmsg(target, argsError)
			}
		case "!clear":
			conn.Privmsg(target, "unimplemented")
		case "!delete":
			conn.Privmsgf(target, "unimplemented")
		}
	})

	watchForInterrupt(conn, db.conn, conf)

	if err := conn.Connect(host); err != nil {
		log.Fatalf("Error connecting: %s\n", err.Error())
	}

	conn.Loop()
}

func buildConf() *Config {
	viper.SetConfigName("dungeonbot")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}

	chanWhole := viper.GetString("chans")
	chanSep := strings.Split(chanWhole, ",")

	return &Config{
		debug:       viper.GetBool("debug_mode"),
		nick:        viper.GetString("nick"),
		user:        viper.GetString("user"),
		chans:       chanSep,
		server:      viper.GetString("server"),
		port:        viper.GetUint("port"),
		tls:         viper.GetBool("tls"),
		pastebinURL: viper.GetString("pastebin_url"),
		dbLocation:  viper.GetString("database_location"),
		signOff:     viper.GetString("signoff"),
	}
}

func watchForInterrupt(conn *irc.Connection, db *sql.DB, conf *Config) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("\n\nCaught %v\n", sigint)
			for _, e := range conf.chans {
				conn.Privmsg(e, conf.signOff)
			}
			conn.Quit()

			if err := db.Close(); err != nil {
				log.Printf("Error closing database connection: %s", err.Error())
			}

			time.Sleep(150 * time.Millisecond)
			os.Exit(1)
		}
	}()
}
