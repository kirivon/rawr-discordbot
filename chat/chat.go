package chat

import (
	"fmt"
	"log"
	"time"

	"github.com/albert-wang/tracederror"
	"github.com/bwmarrin/discordgo"
)

var client *discordgo.Session

// ConnectToWebsocket connects to the discord websocket with the given token.
// This makes the bot appear online, and will begin receiving messages.
func ConnectToWebsocket(token string, onMessage func(*discordgo.Session, *discordgo.MessageCreate)) error {
	var err error
	token = fmt.Sprintf("Bot %s", token)
	client, err = discordgo.New(token)
	if err != nil {
		log.Print("Failed to create discord client")
		return tracederror.New(err)
	}

	client.AddHandler(onMessage)

	err = client.Open()
	if err != nil {
		log.Print("Failed to open connection to discord websocket API")
		return tracederror.New(err)
	}

	return nil
}

func GetChannelInformation(channelID string) (*discordgo.Channel, error) {
	return client.Channel(channelID)
}

// SendMessageToChannel sends a message to a channelID.
func SendMessageToChannel(channelID string, message string) {
	_, err := client.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Print(err)
	}
}

func SendPrivateMessageTo(user string, message string) {
	ch, err := client.UserChannelCreate(user)
	if err != nil {
		log.Print(err)
	}

	SendMessageToChannel(ch.ID, message)
}

// ShowTypingUntilChannelIsClosed will display the bot as typing something in
// the given channel, until a signal is pushed into the golang channel `ch`. Closing
// the golang channel will also stop the typing animation.
func ShowTypingUntilChannelIsClosed(channelID string, ch chan int) {
	t := time.Tick(time.Second / 2 * 5)
	processing := true

	for processing {
		client.ChannelTyping(channelID)
		select {
		case <-t:
			break
		case <-ch:
			processing = false
			break
		}
	}
}
