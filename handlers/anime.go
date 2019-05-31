package handlers

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kirivon/rawr-discordbot/chat"
	"github.com/kirivon/rawr-discordbot/config"
)

//Defines types used in AnimeStatus
type animeStatus struct {
	Name           string
	CurrentEpisode int64
	Members        map[string]string
	LastModified   time.Time
}

//Defines types used in JunbiOK and JunbiRdy
type junbiStatus struct {
	Initialized bool
	Members     map[string]string
}

func (a *animeStatus) FormattedTime() string {
	return a.LastModified.Format("Mon, January 02")
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
	config.LoadConfigFromFileAndENV("config.json")

	//Message user with list of commands if no command is specified
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime <add|drop|del|incr|decr|set|rename|get|list|start> <name> [<value>]")
		return nil
	}

	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Read values from the Redis database, creates key
	key := makeKey("animestatus")
	res := map[string]animeStatus{}
	usr := map[string]string{}
	deserialize(conn, key, &res)

	// Supports add, drop, del, incr, decr, set, rename, get, list, start
	switch args[0] {
	case "add":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime add <name>")
				return nil
			}

			//Checks to see if specified anime exists, adds new entry if it does not
			v, ok := res[args[1]]

			if !ok {
				usr[m.Author.ID] = m.Author.Username
				res[args[1]] = animeStatus{args[1], 0, usr, time.Now()}
				chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Added %s.", args[1]))
				//Upadates Legacy Bot's list
				//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime set %s 0", args[1]))
			} else {
				//Checks to see if the user has already added this anime
				for n, _ := range v.Members {
					if n == m.Author.ID {
						chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("You have already added %s.", args[1]))
						return nil
					}
				}
				//Adds user to Members of specified anime
				v.Members[m.Author.ID] = m.Author.Username
				res[args[1]] = v
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Added %s.", args[1]))
			}
		}
	case "drop":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime drop <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
				return nil
			}

			//Removes user from Members of specified anime
			delete(v.Members, m.Author.ID)
			res[args[1]] = v
			chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Dropped %s.", args[1]))

			//Deletes anime if zero members are present after drop
			if len(v.Members) == 0 {
				delete(res, args[1])
				chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Deleted %s.", args[1]))
				//Upadates Legacy Bot's list
				//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime del %s", args[1]))
			}
		}
	case "del":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime del <name>")
				return nil
			}

			//Deletes specified anime, regardless of members
			delete(res, args[1])
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Deleted %s.", args[1]))
			//Upadates Legacy Bot's list
			//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime del %s", args[1]))
		}
	case "rename":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime rename <name> <new>")
				return nil
			}

			v, ok := res[args[1]]
			_, ok2 := res[args[2]]

			//Checks to see if specified anime exists, aborts if it does not
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
				return nil
			}

			//Checks to see if new desired name exists, aborts if it does
			if ok2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s is already in use, type '.anime list' for a list of shows.", args[2]))
				return nil
			}

			//Changes AnimeStatus.Name of the source element and copies it to the specified target element
			v.Name = args[2]
			res[args[2]] = v
			//Deletes inintial element after copy
			delete(res, args[1])
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Renamed %s to %s.", args[1], args[2]))
			//Upadates Legacy Bot's list
			//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime mv %s %s", args[1], args[2]))
		}
	case "incr", "decr":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("Usage: .anime %s <name>", args[0]))
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
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
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

			if args[0] == "incr" {
				//Upadates Legacy Bot's list
				//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime incr %s", args[1]))
			} else {
				//Upadates Legacy Bot's list
				//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime decr %s", args[1]))
			}
		}
	case "set":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 3 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime set <name> <ep#>")
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
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
				return nil
			}
			//Updates CurrentEpisode and LastModified
			v.CurrentEpisode = episode
			v.LastModified = time.Now()

			//Updates res then sends the new value to chat
			res[args[1]] = v
			chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s - %d (%s)", v.Name, v.CurrentEpisode, v.LastModified.Format("Mon, January 02")))
			//Upadates Legacy Bot's list
			//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime set %s %s", args[1], args[2]))
		}
	case "get":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime get <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
				return nil
			}

			tplText := `Markdown
{{ pad .Len " " "Title" }} | Episode | Last Updated
{{ pad .Len "-" "-----" }}-+---------+-------------
{{ pad $.Len " " .Anime.Name }} | {{ with $x := printf "%d" .Anime.CurrentEpisode }}{{ pad 7 " " $x }}{{ end }} | {{ .Anime.LastModified.Format "Mon, January 02" }} `
			buff := bytes.NewBuffer(nil)

			tpl, err := template.New("anime").Funcs(template.FuncMap{
				"pad": func(amount int, spacer string, val string) string {
					if len(val) < amount {
						return strings.Repeat(spacer, amount-len(val)) + val
					}
					return val
				},
			}).Parse(tplText)

			if err != nil {
				chat.SendMessageToChannel(m.ChannelID, err.Error())
			}

			maximumTitle := 0
			if len(v.Name) > maximumTitle {
				maximumTitle = len(v.Name)
			}

			err = tpl.Execute(buff, map[string]interface{}{
				"Anime": v,
				"Len":   maximumTitle,
			})

			if err != nil {
				log.Print(err)
			}

			message := fmt.Sprintf("\n" + strings.Repeat("-", maximumTitle) + "-+---------+-------------" + "\n")
			message += fmt.Sprintf("Members: ")
			i := 1
			for _, n := range v.Members {
				if len(v.Members) == i {
					message += fmt.Sprintf("%s.", n)
				} else {
					message += fmt.Sprintf("%s, ", n)
				}
				i++
			}

			//Outputs list to chat in codeblock form
			chat.SendMessageToChannel(m.ChannelID, "```"+buff.String()+message+"```")
		}
	case "list":
		{
			//Builds list of existing anime
			tplText := `Markdown
{{ pad .Len " " "Title" }} | Episode | Members | Last Updated
{{ pad .Len "-" "-----" }}-+---------+---------+-------------
{{ range .Animes }}{{ pad $.Len " " .Name }} | {{ with $x := printf "%d" .CurrentEpisode }}{{ pad 7 " " $x }}{{ end }} | {{with $x :=  len .Members | printf "%d" }}{{ pad 7 " " $x }}{{ end }} | {{ .LastModified.Format "Mon, January 02" }}
{{ end }}`

			buff := bytes.NewBuffer(nil)

			tpl, err := template.New("anime").Funcs(template.FuncMap{
				"pad": func(amount int, spacer string, val string) string {
					if len(val) < amount {
						return strings.Repeat(spacer, amount-len(val)) + val
					}
					return val
				},
			}).Parse(tplText)

			if err != nil {
				chat.SendMessageToChannel(m.ChannelID, err.Error())
			}

			maximumTitle := 0
			for _, v := range res {
				if len(v.Name) > maximumTitle {
					maximumTitle = len(v.Name)
				}
			}

			err = tpl.Execute(buff, map[string]interface{}{
				"Animes": res,
				"Len":    maximumTitle,
			})

			if err != nil {
				log.Print(err)
			}

			//Outputs list to chat in codeblock form
			chat.SendMessageToChannel(m.ChannelID, "```"+buff.String()+"```")
		}
	case "start":
		{
			//Sends error to user if there are insufficient arguments
			if len(args) != 2 {
				chat.SendPrivateMessageTo(m.Author.ID, "Usage: .anime start <name>")
				return nil
			}

			//Checks to see if specified anime exists, aborts if it does not
			v, ok := res[args[1]]
			if !ok {
				chat.SendPrivateMessageTo(m.Author.ID, fmt.Sprintf("%s doesn't exist, type '.anime list' for a list of shows.", args[1]))
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
			//Upadates Legacy Bot's list
			//chat.SendPrivateMessageTo(config.AssociatedBotID, fmt.Sprintf("!anime incr %s", args[1]))
		}
	}
	//Write the modified value to the Redis database
	serialize(conn, key, &res)
	return nil
}

//Initializes the values for JunbiRdy (!rdy), sends ready message to channel
func JunbiOK(m *discordgo.MessageCreate, args map[string]string) error {
	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Overwrites any existing value for junbiStatus
	key := makeKey("junbistatus:%s", m.ChannelID)
	res := junbiStatus{true, args}
	serialize(conn, key, &res)

	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("Junbi OK? Type .rdy to confirm!"))
	return nil
}

func JunbiRdy(m *discordgo.MessageCreate, args []string) error {
	//Open connection to Redis server
	conn := Redis.Get()
	defer conn.Close()

	//Read values from the Redis database
	key := makeKey("junbistatus:%s", m.ChannelID)
	res := junbiStatus{}
	deserialize(conn, key, &res)

	//Aborts the function if .anime start hasn't been called
	if res.Initialized != true {
		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("No anime initiated. Type .anime start <name> to begin!"))
		return nil
	}

	//Removes user from junbiStatus.Members
	delete(res.Members, m.Author.ID)

	//Displays the remaining members that haven't confirmed with !rdy
	if len(res.Members) != 0 {
		message := fmt.Sprintf("Waiting on: ")
		i := 1
		for _, n := range res.Members {
			if len(res.Members) == i {
				message += fmt.Sprintf("%s.", n)
			} else {
				message += fmt.Sprintf("%s, ", n)
			}
		}
		chat.SendMessageToChannel(m.ChannelID, message)
		i++
	} else {
		//Resets Initialized flag to false, starts countdown
		res.Initialized = false
		Countdown(m, []string{"3"})
	}

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
