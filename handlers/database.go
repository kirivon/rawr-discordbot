package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"
	"github.com/mitchellh/goamz/s3"
)

type CommandHandler func(*discordgo.MessageCreate, []string) error

var Redis *redis.Pool
var S3Client *s3.S3

func makeKey(f string) string {
	return fmt.Sprintf("rawr-discordbot.%s", f)
}
