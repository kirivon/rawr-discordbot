package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"
)

// RandomS3ImageFrom links a random image from a bucket with the given prefix.
func RandomS3ImageFrom(bucket string, prefix string) CommandHandler {
	return func(m *discordgo.MessageCreate, args []string) error {
		conn := Redis.Get()
		defer conn.Close()

		key := makeKey(fmt.Sprintf("s3rand:%s%s", bucket, prefix))
		contents := []string{}
		res, _ := redis.Bytes(conn.Do("GET", key))
		if res == nil {
			bucket := S3Client.Bucket(bucket)
			resp, err := bucket.List(prefix, "/", "", 1000)
			if err != nil {
				return err
			}

			names := []string{}
			for _, v := range resp.Contents {
				names = append(names, v.Key)
			}

			bytes, err := json.Marshal(names)
			if err != nil {
				return err
			}

			conn.Do("SET", key, string(bytes), "EX", 60*60*24)
			contents = names
		} else {
			err := json.Unmarshal(res, &contents)
			if err != nil {
				return err
			}
		}

		target := rand.Int31n(int32(len(contents)))
		targetKey := contents[target]
		if strings.HasPrefix(targetKey, prefix) {
			targetKey = targetKey[len(prefix):]
		}

		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("http://%s/%s", bucket, prefix+url.QueryEscape(targetKey)))
		return nil
	}
}
