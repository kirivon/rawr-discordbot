package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
)

//Defines types used in AnimeStatus
type animeStatus struct {
	Name           string
	CurrentEpisode int64
	Members        []string
	LastModified   time.Time
}

//Defines types used in JunbiOK and JunbiRdy
type junbiStatus struct {
	Initialized bool
	Members     []string
}

//Constrains passed value to specified range
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
	//Message user with list of commands if no command is specified
	if len(args) < 2 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime <add|drop|del|incr|decr|set|rename|get|list|start> <name> [<value>]")
	}

	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Read values from the Redis database, creates key on per-chat basis
	key := makeKey("animestatus:%s", m.ChannelID)
	res := map[string]animeStatus{}
	deserialize(conn, key, &res)

	// Supports add, drop, del, incr, decr, set, rename, get, list, start
	switch args[0] {
	case "add":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime add <name>")
				return nil
			}

			//Checks to see if specified anime exists, adds new entry if it does not
			v, ok := res[args[1]]
			if !ok {
				res[args[1]] = animeStatus{args[1], 0, []string{m.Author.ID}, time.Now()}
			} else {
				//Checks to see if the user has already added this anime
				for _, n := range v.Members {
					if n == m.Author.ID {
						chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("You have already added %s.", args[1]))
						return nil
					}
				}
				//Adds user to Members
				v.Members = append(v.Members, m.Message.ID)
				res[args[1]] = v
			}
		}
	case "drop":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime drop <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}

			//Removes user from AnimeStatus.Members of specified anime
			i := 0
			for _, n := range v.Members {
				if n != m.Author.ID {
					v.Members[i] = n
					i++
				}
			}
			v.Members = v.Members[:i]
			res[args[1]] = v

			//Deletes anime if zero members are present after drop
			if len(v.Members) == 0 {
				delete(res, args[1])
			}
		}
	case "del":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime del <name>")
				return nil
			}

			//Deletes specified anime, regardless of members
			delete(res, args[1])
		}
	case "rename":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime rename <name> <new>")
				return nil
			}

			v, ok := res[args[1]]
			_, ok2 := res[args[2]]

			//Checks to see if specified anime exists, aborts if it does not
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}

			//Checks to see if new desired name exists, aborts if it does
			if ok2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s is already in use, type '!anime list' for a list of shows.", args[2]))
				return nil
			}

			//Changes AnimeStatus.Name of the source element and copies it to the specified target element
			v.Name = args[2]
			res[args[2]] = v
			//Deletes inintial element after copy
			delete(res, args[1])
		}
	case "incr", "decr":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: !anime %s <name>", args[0]))
				return nil
			}

			//Sets delta to -1 for decr, 1 for incr
			delta := int64(-1)
			if args[0] == "incr" {
				delta = 1
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}

			//Modifies episode count by delta, updates time if incr is called
			v.CurrentEpisode = v.CurrentEpisode + delta
			v.CurrentEpisode = clamp(v.CurrentEpisode, -10, 1000)

			if args[0] == "incr" {
				v.LastModified = time.Now()
			}

			//Updates res then sends the new value to chat
			res[args[1]] = v
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
		}
	case "set":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime set <name> <ep#>")
				return nil
			}

			//Converts arg to int, aborts if err, clamps converted value to specified range
			episode, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}
			episode = clamp(episode, -10, 1000)

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}
			//Updates CurrentEpisode and LastModified
			v.CurrentEpisode = episode
			v.LastModified = time.Now()

			//Updates res then sends the new value to chat
			res[args[1]] = v
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
		}
	case "get":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime get <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}
			//Builds message
			message := "\tName\tLast Ep\tMembers\tLast Watched\n"
			message += fmt.Sprintf("\t%s\t%d\t%d\t%s\n", v.Name, v.CurrentEpisode, len(v.Members), v.LastModified.Format("Mon, January 02"))
			message += fmt.Sprintf("Participants:\t")
			for _, n := range v.Members {
				message += fmt.Sprintf("%s\t", n)
			}
			message += fmt.Sprintf("\n")

			//Outputs message to chat in codeblock form
			chat.SendMessageToChannel(m.ChannelID, "```"+message+"```")
		}
	case "list":
		{
			//Builds list of existing anime
			message := "\tName\tLast Ep\tMembers\tLast Watched\n"
			for _, v := range res {
				message += fmt.Sprintf("\t%s\t%d\t%d\t%s\n", v.Name, v.CurrentEpisode, len(v.Members), v.LastModified.Format("Mon, January 02"))
			}

			//Outputs list to chat in codeblock form
			chat.SendMessageToChannel(m.ChannelID, "```"+message+"```")
		}
	case "start":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: !anime start <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '!anime list' for a list of shows.", args[1]))
				return nil
			}
			//Increments episode and sets time to now
			v.CurrentEpisode++
			v.CurrentEpisode = clamp(v.CurrentEpisode, -10, 1000)
			v.LastModified = time.Now()
			//Updates res then sends the value to chat
			res[args[1]] = v
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Starting %s episode %d.", v.Name, v.CurrentEpisode))
			//Sleeps then calls JunbiOK
			time.Sleep(300 * time.Millisecond)
			JunbiOK(m, v.Members)
		}
	}
	//Write the modified value to the Redis database
	serialize(conn, key, &res)
	return nil
}

//Initializes the values for JunbiRdy (!rdy), sends ready message to channel
func JunbiOK(m *discordgo.MessageCreate, args []string) error {
	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Read values from the Redis database, creates key on per-chat basis
	key := makeKey("junbistatus:%s", m.ChannelID)
	res := junbiStatus{true, args}
	deserialize(conn, key, &res)

	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Junbi OK? Type !rdy to confirm!"))

	//Write the modified value to the Redis database
	serialize(conn, key, &res)
	return nil
}

func JunbiRdy(m *discordgo.MessageCreate, args []string) error {
	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Read values from the Redis database, creates key on per-chat basis
	key := makeKey("junbistatus:%s", m.ChannelID)
	res := junbiStatus{}
	deserialize(conn, key, &res)

	//Aborts the function if !anime start hasn't been called
	if res.Initialized != true {
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("No anime initiated. Type !anime start <name> to begin!"))
		return nil
	}

	//Removes user from junbiStatus.Members
	i := 0
	for _, n := range res.Members {
		if n != m.Author.ID {
			res.Members[i] = n
			i++
		}
	}
	res.Members = res.Members[:i]

	//Displays the remaining members that haven't confirmed with !rdy
	if len(res.Members) != 0 {
		message := fmt.Sprintf("Waiting on:\t")
		for _, n := range res.Members {
			message += fmt.Sprintf("%s\t", n)
		}
		chat.SendMessageToChannel(m.ChannelID, message)
		return nil
	}

	//Resets Initialized flag to false, starts countdown
	res.Initialized = false
	Countdown(m, []string{"3"})

	//Write the modified value to the Redis database
	serialize(conn, key, &res)
	return nil
}

//Simple countdown function
func Countdown(m *discordgo.MessageCreate, args []string) error {
	start := int64(3)
	var err error

	//Sets countdown length to user input, with a fallback value of 3
	if len(args) == 1 {
		start, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			start = 3
		}
	}

	//Constricts countdown length to 30 seconds
	if start > 30 {
		start = 30
	}

	//Starts the countdown
	for i := int64(0); i < start; i++ {
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%d", start-i))
		time.Sleep(time.Second)
	}

	chat.SendMessageToChannel(m.ChannelID, "g")
	return nil
}
