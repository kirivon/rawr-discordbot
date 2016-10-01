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
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/albert-wang/rawr-discordbot/config"
	"github.com/albert-wang/rawr-discordbot/handlers"
)

var mapping map[string]handlers.CommandHandler = map[string]handlers.CommandHandler{}
var argSplit *regexp.Regexp = regexp.MustCompile("'.+'|\".+\"|\\S+")

func help(m *discordgo.MessageCreate, args []string) error {
	msg := "This is NVG-Tan. A listing of commands follows."
	res := []string{}
	for k, _ := range mapping {
		res = append(res, k)
	}

	msg = msg + " " + strings.Join(res, ", ")

	chat.SendMessageToChannel(m.ChannelID, msg)
	return nil
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	handlers.WriteToFile(m, nil)

	args := argSplit.FindAllString(m.Content, -1)
	if len(args) == 0 {
		return
	}

	cmd := args[0]
	args = args[1:]

	if !strings.HasPrefix(cmd, "!") {
		return
	}

	if m.Author.Username == "NVG-Tan" {
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

	auth, err := aws.GetAuth(config.AWSAccessKey, config.AWSSecret)
	if err != nil {
		log.Fatal(err)
	}

	handlers.S3Client = s3.New(auth, aws.USEast)

	// Begin setting up the handlers here
	mapping["help"] = help
	mapping["smug"] = handlers.RandomS3FileFrom("img.rawr.moe", "smug/")
	mapping["kajiura"] = handlers.RandomS3FileFrom("img.rawr.moe", "music/")
	mapping["search"] = handlers.Search
	mapping["search-help"] = handlers.SearchHelp
	mapping["countdown"] = handlers.Countdown
	mapping["anime"] = handlers.AnimeStatus
	mapping["junbiok"] = handlers.JunbiOK
	mapping["rdy"] = handlers.JunbiRdy

	mux := http.NewServeMux()
	mux.HandleFunc("/searchresult", handlers.SearchResults)
	chat.ConnectToWebsocket(config.BotToken, onMessage)

	log.Printf("Listening on :%s", config.InternalBindPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", config.InternalBindPort), mux)
	if err != nil {
		log.Print(err)
	}
}
