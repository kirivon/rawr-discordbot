package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/kirivon/rawr-discordbot/chat"
	"github.com/kirivon/rawr-discordbot/config"
)

type GoogleResult struct {
	Items []struct {
		Value string `json:"link"`
	} `json:"items"`
}

type SafebooruResult struct {
	Items []struct {
		Value string `xml:"file_url,attr"`
	} `xml:"post"`
}

func GoogleSearch(m *discordgo.MessageCreate, args []string) error {
	config.LoadConfigFromFileAndENV("config.json")

	//Sends error to user if there are insufficeint arguments
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage .g <value>")
		return nil
	}

	//Concatenates user search strings if more than one is present
	var searchstring string
	if len(args) > 1 {
		for i, _ := range args {
			searchstring += args[i]
			//Deliniates values with +
			if i != len(args)-1 {
				searchstring += "+"
			}
		}
	} else {
		searchstring = args[0]
	}

	//Utilizes custom Google Search API for the user input and returns first result
	res, err := http.Get(fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&num=1&q=%s", config.GoogleAPIKey, config.SearchEngineID, searchstring))

	//Sends error message to chat
	if err != nil {
		chat.SendMessageToChannel(m.ChannelID, err.Error())
		return nil
	}

	//Closes response body when finished
	defer res.Body.Close()

	//Assigns struct to a variable to store JSON output
	var res2 GoogleResult
	//Reads the HTTP.Get response and Decodes the values to res2
	err2 := json.NewDecoder(res.Body).Decode(&res2)

	//Logs error to console
	if err2 != nil {
		log.Print(err2)
	}

	if len(res2.Items) == 0 {
		chat.SendMessageToChannel(m.ChannelID, "No results found!")
		return nil
	}

	//Sends the link to chat
	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s", res2.Items[0].Value))

	return nil
}

func GoogleImageSearch(m *discordgo.MessageCreate, args []string) error {
	config.LoadConfigFromFileAndENV("config.json")

	//Sends error to user if there are insufficeint arguments
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage .g <value>")
		return nil
	}

	//Concatenates user search strings if more than one is present
	var searchstring string
	if len(args) > 1 {
		for i, _ := range args {
			searchstring += args[i]
			//Deliniates values with +
			if i != len(args)-1 {
				searchstring += "+"
			}
		}
	} else {
		searchstring = args[0]
	}

	//Utilizes custom Google Search API for the user input and returns first result
	res, err := http.Get(fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&num=1&q=%s&searchType=image", config.GoogleAPIKey, config.SearchEngineID, searchstring))

	//Sends error message to chat
	if err != nil {
		chat.SendMessageToChannel(m.ChannelID, err.Error())
		return nil
	}

	//Closes response body when finished
	defer res.Body.Close()

	//Assigns struct to a variable to store JSON output
	var res2 GoogleResult
	//Reads the HTTP.Get response and Decodes the values to res2
	err2 := json.NewDecoder(res.Body).Decode(&res2)

	//Logs error to console
	if err2 != nil {
		log.Print(err2)
	}

	if len(res2.Items) == 0 {
		chat.SendMessageToChannel(m.ChannelID, "No results found!")
		return nil
	}

	//Sends the link to chat
	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s", res2.Items[0].Value))

	return nil
}

func YoutubeSearch(m *discordgo.MessageCreate, args []string) error {
	config.LoadConfigFromFileAndENV("config.json")

	//Sends error to user if there are insufficeint arguments
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage .g <value>")
		return nil
	}

	//Concatenates user search strings if more than one is present
	var searchstring string
	if len(args) > 1 {
		for i, _ := range args {
			searchstring += args[i]
			//Deliniates values with +
			if i != len(args)-1 {
				searchstring += "+"
			}
		}
	} else {
		searchstring = args[0]
	}

	SearchEngineID := "008171425300186654227:cxh4x9ted3a"

	//Utilizes custom Google Search API for the user input and returns first result
	res, err := http.Get(fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&num=1&q=%s", config.GoogleAPIKey, SearchEngineID, searchstring))

	//Sends error message to chat
	if err != nil {
		chat.SendMessageToChannel(m.ChannelID, err.Error())
		return nil
	}

	//Closes response body when finished
	defer res.Body.Close()

	//Assigns struct to a variable to store JSON output
	var res2 GoogleResult
	//Reads the HTTP.Get response and Decodes the values to res2
	err2 := json.NewDecoder(res.Body).Decode(&res2)

	//Logs error to console
	if err2 != nil {
		log.Print(err2)
	}

	if len(res2.Items) == 0 {
		chat.SendMessageToChannel(m.ChannelID, "No results found!")
		return nil
	}

	//Sends the link to chat
	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s", res2.Items[0].Value))

	return nil
}

func SafebooruSearch(m *discordgo.MessageCreate, args []string) error {

	//Sends error to user if there are insufficeint arguments
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage .g <value>")
		return nil
	}

	//Concatenates user search strings if more than one is present
	var searchstring string
	if len(args) > 1 {
		for i, _ := range args {
			searchstring += args[i]
			//Deliniates values with +
			if i != len(args)-1 {
				searchstring += "+"
			}
		}
	} else {
		searchstring = args[0]
	}

	//Utilizes custom Google Search API for the user input and returns first result
	res, err := http.Get(fmt.Sprintf("http://safebooru.org/index.php?page=dapi&s=post&q=index&limit=1&tags=%s", searchstring))

	//Sends error message to chat
	if err != nil {
		chat.SendMessageToChannel(m.ChannelID, err.Error())
		return nil
	}

	//Closes response body when finished
	defer res.Body.Close()

	//Assigns struct to a variable to store JSON output
	var res2 SafebooruResult
	//Reads the HTTP.Get response and Decodes the values to res2
	err2 := xml.NewDecoder(res.Body).Decode(&res2)

	//Logs error to console
	if err2 != nil {
		log.Print(err2)
	}

	if len(res2.Items) == 0 {
		chat.SendMessageToChannel(m.ChannelID, "No results found!")
		return nil
	}

	//Sends the link to chat
	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s", res2.Items[0].Value))

	return nil
}

func MTGSearch(m *discordgo.MessageCreate, args []string) error {
	config.LoadConfigFromFileAndENV("config.json")

	//Sends error to user if there are insufficeint arguments
	if len(args) < 1 {
		chat.SendPrivateMessageTo(m.Author.ID, "Usage .g <value>")
		return nil
	}

	//Concatenates user search strings if more than one is present
	var searchstring string
	if len(args) > 1 {
		for i, _ := range args {
			searchstring += args[i]
			//Deliniates values with +
			if i != len(args)-1 {
				searchstring += "+"
			}
		}
	} else {
		searchstring = args[0]
	}

	SearchEngineID := "008171425300186654227:umenevadeny"

	//Utilizes custom Google Search API for the user input and returns first result
	res, err := http.Get(fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&num=1&q=%s&searchType=image", config.GoogleAPIKey, SearchEngineID, searchstring))

	//Sends error message to chat
	if err != nil {
		chat.SendMessageToChannel(m.ChannelID, err.Error())
		return nil
	}

	//Closes response body when finished
	defer res.Body.Close()

	//Assigns struct to a variable to store JSON output
	var res2 GoogleResult
	//Reads the HTTP.Get response and Decodes the values to res2
	err2 := json.NewDecoder(res.Body).Decode(&res2)

	//Logs error to console
	if err2 != nil {
		log.Print(err2)
	}

	if len(res2.Items) == 0 {
		chat.SendMessageToChannel(m.ChannelID, "No results found!")
		return nil
	}

	//Sends the link to chat
	chat.SendMessageToChannel(m.ChannelID, fmt.Sprintf("%s", res2.Items[0].Value))

	return nil
}
