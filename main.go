package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/albert-wang/rawr-discordbot/config"
	"github.com/albert-wang/rawr-discordbot/handlers"
)

type CommandHandler func(*discordgo.MessageCreate, []string) error;

var mapping map[string]CommandHandler = map[string]CommandHandler{}

func help(m *discordgo.MessageCreate, args []string) {
	msg := "This is NVG-Tan. A listing of commands follows.";
	res := []string{}
	for k, _ := range mapping {
		res = append(res, k)
	}

	msg = msg + " " + strings.Join(res, ", ")

	chat.SendMessageToChannel(m.ChannelID, msg)
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 0 {
		return 
	}

	cmd := args[0]
	args = args[1:]

	if !strings.HasPrefix(cmd, "!") {
		return
	}

	handler, ok := mapping[cmd[1:]]
	if ok {
		handler(m, args)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	config.LoadConfigFromFileAndENV("config.json")

	handlers.Redis = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.RedisServerAddress)
			if err != nil {
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// Begin setting up the handlers here
	mapping["help"] = help



	mux := http.NewServeMux()
	chat.ConnectToWebsocket(config.BotToken, onMessage)

	log.Printf("Listening on :%s", config.InternalBindPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", config.InternalBindPort), mux)
	if err != nil {
		log.Print(err)
	}
}
