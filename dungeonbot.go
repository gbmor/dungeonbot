package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

// VERSION is the dungeonbot version
var VERSION = ""

// Config holds deserialized data from dungeonbot.yml
type Config struct {
	nick   string
	user   string
	chans  []string
	server string
	port   uint
	ssl    bool
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
	conn.Debug = true
	conn.UseTLS = conf.ssl
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: false}

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
			out, err := parseDice(msg[1])
			if err != nil {
				conn.Privmsgf(target, "%s", err.Error())
				return
			}
			conn.Privmsgf(target, "%s", out)
		}
	})

	watchForInterrupt(conn, conf.nick)

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
		nick:   viper.GetString("nick"),
		user:   viper.GetString("user"),
		chans:  chanSep,
		server: viper.GetString("server"),
		port:   viper.GetUint("port"),
		ssl:    viper.GetBool("ssl"),
	}
}

func watchForInterrupt(conn *irc.Connection, nick string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("Caught %v\n", sigint)
			conn.SendRawf("QUIT /me yeet %s", nick)
			time.Sleep(50 * time.Millisecond)
			os.Exit(0)
		}
	}()
}
