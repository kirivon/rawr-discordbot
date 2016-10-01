package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
)

type animeStatus struct {
	Name           string
	CurrentEpisode int64
	LastModified   time.Time
}

func clamp(v, l, h int64) int64 {
	if v < l {
		return l
	}

	if v > h {
		return h
	}

	return v
}

func AnimeStatus(m *discordgo.MessageCreate, args []string) error {
	if len(args) < 2 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime <del|mv|incr|decr|set|list|get|start> <name> [<value>]")
	}

	conn := Redis.Get()
	defer conn.Close()

	key := makeKey("animestatus:%s", m.Author.ID)
	res := map[string]animeStatus{}
	deserialize(conn, key, &res)

	// Supports del, mv, incr, decr, set, list, start
	switch args[0] {
	case "del":
		{
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime del <name>")
				return nil
			}

			delete(res, args[1])
			break
		}
	case "mv":
		{
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime mv <name> <new>")
				return nil
			}

			_, ok := res[args[2]]
			v, ok2 := res[args[1]]

			if ok || !ok2 {
				chat.SendPrivateMessageTo(m.Author.ID, "!anime mv cannot overwrite elements, or source element did not exist")
			}

			v.Name = args[2]
			res[args[2]] = v
			delete(res, args[1])
			break
		}
	case "set":
		{
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime set <name> <ep#>")
				return nil
			}

			episode, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			episode = clamp(episode, -10, 1000)
			v, ok := res[args[1]]
			if !ok {
				res[args[1]] = animeStatus{args[1], episode, time.Now()}
			} else {
				v.CurrentEpisode = episode
				v.LastModified = time.Now()
				res[args[1]] = v
			}

			v = res[args[1]]
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
			break
		}
	case "get":
		{
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime get <name>")
				return nil
			}

			v, ok := res[args[1]]
			if !ok {
				chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s doesn't exist, try list", args[1]))
				return nil
			}
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
		}
	case "incr", "decr":
		{
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: !anime %s <name>", args[0]))
				return nil
			}

			delta := int64(-1)
			if args[0] == "incr" {
				delta = 1
			}

			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: !anime %s <name> requires a valid name", args[0]))
				return nil
			} else {
				v.CurrentEpisode = v.CurrentEpisode + delta
				v.CurrentEpisode = clamp(v.CurrentEpisode, -10, 1000)

				if args[0] == "incr" {
					v.LastModified = time.Now()
				}

				res[args[1]] = v
				chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
			}
		}
	case "list":
		{
			message := ""
			for _, v := range res {
				message += fmt.Sprintf("\t%s\t%d\t%s\n", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02"))
			}

			chat.SendMessageToChannel(m.ChannelID, "```"+message+"```")
		}
	case "start":
		{
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: !anime %s <name>", args[0]))
				return nil
			}

			delta := int64(1)
			v, ok := res[args[1]]

			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: !anime %s <name> requires a valid name", args[0]))
				return nil
			} else {
				v.CurrentEpisode = v.CurrentEpisode + delta
				v.CurrentEpisode = clamp(v.CurrentEpisode, -10, 1000)
				v.LastModified = time.Now()
				res[args[1]] = v
				chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Starting %s episode %d.", v.Name, v.CurrentEpisode))
				time.Sleep(300 * time.Millisecond)
				JunbiOK(m, []string{"3"})
			}
		}
	}

	serialize(conn, key, &res)
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

var (
	junbiCount, junbiMembers int64
	junbiInitiated           bool
)

func JunbiOK(m *discordgo.MessageCreate, args []string) error {
	var err error
	junbiCount = 1
	junbiMembers = 3
	junbiInitiated = true

	if len(args) == 1 {
		junbiMembers, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			junbiMembers = 3
		}
	}

	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Junbi OK? Type !rdy to confirm!"))
	return nil
}

func JunbiRdy(m *discordgo.MessageCreate, args []string) error {
	if junbiInitiated != true {
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Countdown has not been initiated! Type !junbiok to begin!"))
		return nil
	} else {
		junbiCount++
	}

	if junbiCount < junbiMembers {
		count := int64(junbiMembers - junbiCount)
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Waiting on %d more!", count))
		return nil
	} else {
		Countdown(m, []string{"3"})
		junbiInitiated = false
	}
	return nil
}
