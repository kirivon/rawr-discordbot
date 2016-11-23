package handlers

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func AnimeStatusLegacy(m *discordgo.MessageCreate, args []string) error {
	//Prevents runtime panic if no arguments are specified
	if len(args) < 1 {
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

	//Supports legacy set (add), del, mv (rename), incr, decr
	switch args[0] {
	case "set":
		{
			//Aborts if there are insufficient arguments
			if len(args) != 3 {
				return nil
			}

			//Converts arg to int, aborts if err, clamps converted value to specified range
			episode, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}
			episode = clamp(episode, -10, 1000)

			//Checks to see if specified anime exists, adds new entry if it does not
			v, ok := res[args[1]]

			if !ok {
				usr[m.Author.ID] = m.Author.Username
				res[args[1]] = animeStatus{args[1], episode, usr, time.Now()}

			} else {
				v.CurrentEpisode = episode
				v.LastModified = time.Now()
				res[args[1]] = v
			}

		}
	case "del":
		{
			//Aborts if there are insufficient arguments
			if len(args) != 2 {
				return nil
			}

			//Deletes specified anime, regardless of members
			delete(res, args[1])
		}
	case "mv":
		{
			//Aborts if there are insufficient arguments
			if len(args) != 3 {
				return nil
			}

			v, ok := res[args[1]]
			_, ok2 := res[args[2]]

			//Checks to see if specified anime exists, aborts if it does not
			if !ok {
				return nil
			}

			//Checks to see if new desired name exists, aborts if it does
			if ok2 {
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
			//Aborts if there are insufficient arguments
			if len(args) != 2 {
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
				return nil
			}

			//Modifies episode count by delta, updates time if incr is called
			v.CurrentEpisode = v.CurrentEpisode + delta
			v.CurrentEpisode = clamp(v.CurrentEpisode, -10, 1000)

			if args[0] == "incr" {
				v.LastModified = time.Now()
			}

			//Updates res
			res[args[1]] = v
		}
	}
	//Write the modified value to the Redis database
	serialize(conn, key, &res)
	return nil
}
