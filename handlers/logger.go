package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"
)

func WriteToFile(m *discordgo.MessageCreate, args []string) error {
	ch, err := chat.GetChannelInformation(m.ChannelID)
	if err != nil {
		return err
	}

	if !ch.IsPrivate {
		conn := Redis.Get()
		defer conn.Close()

		files := []string{}

		for _, v := range m.Attachments {
			files = append(files, v.URL)
		}

		fileMsg := ""
		if len(files) > 0 {
			fileMsg = "\x01" + strings.Join(files, "\x01")
		}

		conn.Do("ZADD", makeKey("chatlog"), time.Now().UTC().Unix(), fmt.Sprintf("%d\x01%s\x01%s%s",
			time.Now().UTC().Unix(), m.Author.Username, m.Content, fileMsg))

		// Only store the past year worth of data.
		conn.Do("ZREMRANGEBYSCORE", makeKey("chatlog"), 0, time.Now().UTC().Add(-1*time.Hour*24*356).Unix())
	}
	return nil
}

func SearchHelp(m *discordgo.MessageCreate, args []string) error {
	ch, err := chat.GetChannelInformation(m.ChannelID)
	if err != nil {
		return err
	}

	if !ch.IsPrivate {
		chat.SendPrivateMessageTo(m.Author.ID, "Please search through PMs to NVG-Tan only! Use !search-help for help with search syntax.")
		return nil
	}

	chat.SendPrivateMessageTo(m.Author.ID, "Syntax: !search [between 'ts' and 'ts'] [username said] [with <number> context] <regex>")
	return nil
}

func Search(m *discordgo.MessageCreate, args []string) error {
	ch, err := chat.GetChannelInformation(m.ChannelID)
	if err != nil {
		return err
	}

	if !ch.IsPrivate {
		chat.SendPrivateMessageTo(m.Author.ID, "Please search through PMs to NVG-Tan only")
		return nil
	}

	// 10 searches every minute.
	conn := Redis.Get()
	defer conn.Close()

	key := makeKey("searchlimit:%s", m.Author.ID)
	v, err := redis.Int(conn.Do("INCR", key))
	if err != nil {
		return err
	}

	if v == 1 {
		conn.Do("EXPIRE", key, 60)
	}

	if v > 10 {
		chat.SendPrivateMessageTo(m.Author.ID, "Searches are limited to 10 every minute. Please wait a bit before searching again")
		return nil
	}

	end := time.Now().UTC().Unix()
	start := time.Now().UTC().Add(-1 * time.Hour * 24 * 30).Unix()

	//Syntax: !search [between 'ts' and 'ts'] [username said] [with <n> context] <regex>
	if len(args) > 4 && args[0] == "between" {
		// Try to consume, if fail, try to use 'between' as a username.
		if args[2] == "and" {
			begin, err := time.Parse("2006-01-02 15:04:05 MST", args[1])
			if err != nil {
				begin, err = time.Parse("2006-01-02 15:04:05", args[1])
			}

			if err != nil {
				chat.SendMessageToChannel(m.Author.ID, "Date format is 2006-01-02 15:04:05 [MST]")
				return nil
			}

			start = begin.Unix()

			e, err := time.Parse("2006-01-02 15-04-05 MST", args[3])
			if err != nil {
				e, err = time.Parse("2006-01-02 15-04-05", args[3])
			}

			if err != nil {
				chat.SendMessageToChannel(m.Author.ID, "Date format is 2006-01-02 15:04:05 [MST]")
				return nil
			}

			end = e.Unix()
		}

		args = args[4:]
	}

	username := ""
	if len(args) > 2 && args[1] == "said" {
		username = args[0]
		args = args[2:]
	}

	context := int64(3)
	if len(args) > 3 && args[0] == "with" && args[2] == "context" {
		context, _ = strconv.ParseInt(args[1], 10, 64)
		args = args[3:]
	}

	if context < 3 {
		context = 3
	}

	for i := range args {
		if strings.HasPrefix(args[i], "\"") && strings.HasSuffix(args[i], "\"") {
			args[i] = args[i][1 : len(args)-2]
		}
	}

	reg := strings.Join(args, " ")
	r, err := regexp.Compile("(?ims)" + reg)

	if err != nil {
		chat.SendPrivateMessageTo(m.Author.ID, "Couldn't compile regex")
		chat.SendPrivateMessageTo(m.Author.ID, err.Error())
		return nil
	}

	strs, err := redis.Strings(conn.Do("ZRANGEBYSCORE", makeKey("chatlog"), start, end))
	if err != nil {
		log.Print(err)
		return err
	}

	before := []string{}
	result := []string{}
	matches := 0

	afterCount := int64(-1)
	for _, v := range strs {
		if afterCount == context {
			chat.SendPrivateMessageTo(m.Author.ID, "```"+strings.Join(before, "\n")+"\n"+strings.Join(result, "\n")+"````")
			before = []string{}
			result = []string{}
			afterCount = int64(-1)
		}

		tsUsernameMessage := strings.Split(v, "\x01")
		ts, _ := strconv.ParseInt(tsUsernameMessage[0], 10, 64)

		searchableMessage := tsUsernameMessage[2]
		if len(tsUsernameMessage) > 3 {
			for _, v := range tsUsernameMessage[3:] {
				searchableMessage = searchableMessage + fmt.Sprintf(" [With uploaded file: %s]", v)
			}
		}

		searchableMessage = strings.Replace(searchableMessage, "\n", "\n\t", -1)
		msg := fmt.Sprintf("[%s] %s: %s", time.Unix(ts, 0).String(), tsUsernameMessage[1], searchableMessage)

		if afterCount < 0 && context > 0 {
			before = append(before, msg)
			if int64(len(before)) >= context {
				before = before[:len(before)-1]
			}
		}

		if afterCount >= 0 && afterCount < context {
			result = append(result, msg)
			afterCount++
		}

		if len(username) > 0 && !strings.EqualFold(username, tsUsernameMessage[1]) {
			continue
		}

		if r.Match([]byte(searchableMessage)) && matches < 16 {
			if len(before) > 0 {
				before = before[1:]
			}

			if afterCount < 0 {
				result = append(result, msg)
			}

			afterCount = 0
		}
	}

	if afterCount >= 0 {
		log.Print(before)
		log.Print(result)
		chat.SendPrivateMessageTo(m.Author.ID, "```"+strings.Join(before, "\n")+"\n"+strings.Join(result, "\n")+"````")
	}

	return nil
}
