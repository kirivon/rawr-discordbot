package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
)

type animeStatus struct {
	CurrentEpisode int
	LastWatched    time.Time
}

func AnimeStatus(m *discordgo.MessageCreate, args []string) error {
	if len(args) != 1 {
		chat.SendMessageToChannel(m.ChannelID, "What anime do you want the status of?")
	}

	conn := Redis.Get()
	defer conn.Close()

	return nil
}

func Countdown(m *discordgo.MessageCreate, args []string) error {
	start := int64(3)
	var err error

	if len(args) == 1 {
		start, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			start = 3
		}
	}

	if start > 30 {
		start = 30
	}

	for i := int64(0); i < start; i++ {
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%d", start-i))
		time.Sleep(time.Second)
	}

	chat.SendMessageToChannel(m.ChannelID, "g")
	return nil
}
