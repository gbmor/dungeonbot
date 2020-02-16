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
	db.init()

	conn.AddCallback("001", func(e *irc.Event) {
		for i := 0; i < len(conf.chans); i++ {
			conn.Join(conf.chans[i])
		}
	})

	conn.AddCallback("*", func(e *irc.Event) {
		splitRaw := strings.Split(e.Raw, " ")
		if splitRaw[0] == "PING" {
			e.Connection.SendRawf("PONG %s", splitRaw[1])
			return
		}
		target := splitRaw[2]

		if strings.HasPrefix(e.Message(), "rain drop") {
			conn.Privmsg(target, "drop top")
			return
		}

		msg := strings.Split(e.Message(), " ")
		switch msg[0] {
		case "!roll":
			if len(msg) < 2 {
				conn.Privmsg(target, "Missing dice argument. Eg: !roll 1d20")
				return
			}

			out, err := parseDice(msg[1])
			if err != nil {
				conn.Privmsgf(target, "%s", err.Error())
				return
			}

			conn.Privmsgf(target, "%s", out)

		case "!campaign":
			if len(msg) < 2 {
				conn.Privmsgf(target, "Missing campaign name. Eg: !campaign gronkulousness")
				return
			}

			arg := strings.Join(msg[1:], " ")
			conn.Privmsgf(target, "Looking for %s campaign notes...", arg)

			raw, err := db.getCampaignNotes(arg)
			if err != nil {
				conn.Privmsgf(target, "No campaign notes for %s", arg)
				log.Printf("%s", err.Error())
				return
			}

			pbURL, err := pastebin(conf.pastebinURL, raw)
			if err != nil {
				conn.Privmsgf(target, "Error connecting to pastebin service")
				log.Printf("%s", err.Error())
				return
			}

			conn.Privmsgf(target, "%s", pbURL)

		case "!add":
			if len(msg) < 2 {
				conn.Privmsgf(target, "Missing subcommand: campaign|pc|npc|monster")
				return
			}
			if len(msg) < 3 {
				conn.Privmsgf(target, "Missing argument. Eg: !add campaign gronkulousness")
				return
			}
			subcommand := msg[1]

			switch subcommand {
			case "campaign":
			case "pc":
			case "npc":
			case "monster":
			}

		case "!append":
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
			os.Exit(0)
		}
	}()
}
