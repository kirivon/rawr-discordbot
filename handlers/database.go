package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"
	"github.com/mitchellh/goamz/s3"
)

type CommandHandler func(*discordgo.MessageCreate, []string) error

var Redis *redis.Pool
var S3Client *s3.S3

func makeKey(f string, args ...interface{}) string {
	return fmt.Sprintf("rawr-discordbot.%s", fmt.Sprintf(f, args...))
}

func cached(key string, timeout int, out interface{}, gen func() (interface{}, error)) error {
	conn := Redis.Get()
	defer conn.Close()

	bytes, err := redis.Bytes(conn.Do("GET", key))
	if bytes == nil {
		res, err := gen()
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(res)
		if err != nil {
			return err
		}

		_, err = conn.Do("SET", string(encoded), "EX", timeout)
		if err != nil {
			return err
		}

		return json.Unmarshal(encoded, out)
	} else {
		if err != nil {
			return err
		}

		return json.Unmarshal(bytes, out)
	}
}
