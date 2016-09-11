package handlers

import (
	"fmt"
	"math/rand"

	"github.com/albert-wang/rawr-discordbot/chat"
	"github.com/bwmarrin/discordgo"
)

// RandomS3ImageFrom links a random image from a bucket with the given prefix.
func RandomS3ImageFrom(bucket string, prefix string) CommandHandler {
	return func(m *discordgo.MessageCreate, args []string) error {
		bucket := S3Client.Bucket(bucket)
		resp, err := bucket.List(prefix, "/", "", 1000)
		if err != nil {
			return err
		}

		target := rand.Int31n(int32(len(resp.Contents)))
		key := resp.Contents[target]

		chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("https://s3.amazonaws.com/%s/%s", bucket, key.Key))
	}
}
