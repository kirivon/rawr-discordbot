package handlers

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
)

// RandomS3FileFrom links a random file from a bucket with the given prefix.
func RandomS3FileFrom(bucket string, prefix string) CommandHandler {
	return func(m *discordgo.MessageCreate, args []string) error {
		conn := Redis.Get()
		defer conn.Close()

		key := makeKey("s3rand:%s%s", bucket, prefix)
		contents := []string{}

		err := cached(key, 60*60*24, &contents, func() (interface{}, error) {
			bucket := S3Client.Bucket(bucket)
			resp, err := bucket.List(prefix, "/", "", 1000)
			if err != nil {
				return nil, err
			}

			names := []string{}
			for _, v := range resp.Contents {
				names = append(names, v.Key)
			}
			return names, nil
		})

		if err != nil {
			return err
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
