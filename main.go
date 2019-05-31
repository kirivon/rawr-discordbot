package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"

	"github.com/kirivon/rawr-discordbot/chat"
	"github.com/kirivon/rawr-discordbot/config"
	"github.com/kirivon/rawr-discordbot/handlers"
)

var mapping map[string]handlers.CommandHandler = map[string]handlers.CommandHandler{}
var argSplit *regexp.Regexp = regexp.MustCompile("'.+'|\".+\"|\\S+")

func help(m *discordgo.MessageCreate, args []string) error {
	msg := "This is Yuno-tan. A listing of commands follows:\n"
	msg += ".anime <add|drop|del|incr|decr|set|rename|get|list|start> <name> [<value>]\n"
	msg += "Searches: Google (.g or .google), Google Image (.gi), Youtube (.yt), Safebooru (.sf), MTG (.mtg)"

	chat.SendMessageToChannel(m.ChannelID, msg)
	return nil
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	//handlers.WriteToFile(m, nil)

	args := argSplit.FindAllString(m.Content, -1)
	if len(args) == 0 {
		return
	}

	cmd := args[0]
	args = args[1:]

	if m.Author.Username == "Yuno-tan" {
		return
	}

	if cmd == "!anime" {
		handler, ok := mapping[cmd]
		if ok {
			go func() {
				err := handler(m, args)
				if err != nil {
					log.Print(err)
				}
			}()
		}
		return
	}

	if !strings.HasPrefix(cmd, ".") {
		return
	}

	handler, ok := mapping[cmd[1:]]
	if ok {
		go func() {
			err := handler(m, args)
			if err != nil {
				log.Print(err)
			}
		}()
		return
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().Unix())

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
	mapping["anime"] = handlers.AnimeStatus
	mapping["!anime"] = handlers.AnimeStatusLegacy
	mapping["rdy"] = handlers.JunbiRdy
	mapping["g"] = handlers.GoogleSearch
	mapping["google"] = handlers.GoogleSearch
	mapping["gi"] = handlers.GoogleImageSearch
	mapping["yt"] = handlers.YoutubeSearch
	mapping["sf"] = handlers.SafebooruSearch
	mapping["mtg"] = handlers.MTGSearch

	mux := http.NewServeMux()
	//mux.HandleFunc("/searchresult", handlers.SearchResults)
	chat.ConnectToWebsocket(config.BotToken, onMessage)

	//	ticker := time.NewTicker(time.Second * 5)
	//	go func() {
	//		for range ticker.C {
	//			log.Print(time.Now().UTC().Unix())
	//		}
	//	}()

	log.Printf("Listening on :%s", config.InternalBindPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", config.InternalBindPort), mux)
	if err != nil {
		log.Print(err)
	}
}
