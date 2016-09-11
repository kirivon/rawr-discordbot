package handlers

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

var Redis *redis.Pool

func makeKey(f string) string {
	return fmt.Sprintf("rawr-discordbot.%s", f)
}
